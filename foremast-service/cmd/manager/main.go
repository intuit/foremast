package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	common "foremast.ai/foremast/foremast-service/pkg/common"
	converter "foremast.ai/foremast/foremast-service/pkg/converter"
	models "foremast.ai/foremast/foremast-service/pkg/models"
	prometheus "foremast.ai/foremast/foremast-service/pkg/prometheus"
	search "foremast.ai/foremast/foremast-service/pkg/search"
	wavefront "foremast.ai/foremast/foremast-service/pkg/wavefront"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
)

var (
	elasticClient *elastic.Client
)

// QueryEndpoint query service endpoint
var QueryEndpoint string

// ConfigSeparator .... constant variable based on to separate the queries
const ConfigSeparator = " ||"

// KvSeparator   .... used for key and value separate
const KvSeparator = "== "

func constructURL(metricQuery models.MetricQuery) (int32, string, string) {
	config := metricQuery.Parameters
	if config == nil || len(config) == 0 {
		return 404, "", ""
	}

	if metricQuery.DataSourceType == "prometheus" {
		return 0, "prometheus", prometheus.BuildURL(metricQuery)
	}
	if metricQuery.DataSourceType == "wavefront" {
		return 0, "wavefront", wavefront.BuildURL(metricQuery)
	}
	//type is not supported
	return 404, metricQuery.DataSourceType, ""
}

func convertMetricQuerys(metric map[string]models.MetricQuery, strategy string) (int32, string, string) {
	if len(metric) == 0 {
		return 404, "", ""
	}
	output := strings.Builder{}
	metricSourceOutput := strings.Builder{}
	var co int

	for key, value := range metric {
		if "hpa" == strategy || "continuous" == strategy {
			log.Printf("parameters %#v", value.Parameters)
			value.Parameters["start"] = "START_TIME"
			value.Parameters["end"] = "END_TIME"
		}
		errCode, metricSource, retstr := constructURL(value)
		if errCode != 0 {
			return 404, retstr, metricSource
		}
		if co == 1 {
			output.WriteString(ConfigSeparator)
			metricSourceOutput.WriteString(ConfigSeparator)
		}
		output.WriteString(key)
		metricSourceOutput.WriteString(key)
		output.WriteString(KvSeparator)
		metricSourceOutput.WriteString(KvSeparator)
		output.WriteString(retstr)
		metricSourceOutput.WriteString(metricSource)
		co = 1
	}
	return 0, output.String(), metricSourceOutput.String()
}

func convertMetricInfoString(m models.MetricsInfo, strategy string) (int, string, []string, []string, map[string]models.HPAMetric) {

	configs := []string{"", "", ""}
	mSources := []string{"", "", ""}
	hpametrics := map[string]models.HPAMetric{}
	if m.Current == nil || len(m.Current) == 0 {
		return 404, "MetricInfo current is empty ", configs, mSources, hpametrics
	}
	errorCode := 0
	reason := strings.Builder{}
	errCode, ret, mSource := convertMetricQuerys(m.Current, strategy)

	if errCode != 0 {
		log.Println("Error: current convertMetricQuerys ", m.Current, " failed. errorCode is ", errCode)
		reason.WriteString("current query encount error ")
		reason.WriteString(ret)
		reason.WriteString("\n")
		errorCode = 404
	}
	configs[0] = ret
	mSources[0] = mSource

	if m.Baseline != nil {
		errCode, ret, mSource := convertMetricQuerys(m.Baseline, strategy)
		if errCode != 0 {
			log.Println("Warning: baseline convertMetricQuerys ", m.Baseline, " failed. errorCode is ", errCode)
			reason.WriteString(" baseline query encount error ")
			reason.WriteString(ret)
		}
		configs[1] = ret
		mSources[1] = mSource
	}

	if m.Historical != nil {
		hErrCode, ret, mSource := convertMetricQuerys(m.Historical, strategy)
		if strategy == "hpa" {
			for k, v := range m.Historical {
				priority := 1
				if v.Priority != nil {
					priority = *v.Priority
				}
				hpametrics[k] = models.HPAMetric{Priority: priority, IsIncrease: v.IsIncrease, IsAbsolute: v.IsAbsolute}
			}
		}
		if hErrCode != 0 {
			log.Println("Warning: historical convertMetricQuerys ", m.Historical, " failed. errorCode is ", hErrCode)
			reason.WriteString(" historical query encount error ")
			reason.WriteString(ret)
		}
		if errCode != 0 && hErrCode != 0 {
			errorCode = 404
		}
		configs[2] = ret
		mSources[2] = mSource
	} else {
		if errCode != 0 {
			errorCode = 404
		}
	}
	if strategy != "hpa" {
		return errorCode, reason.String(), configs, mSources, nil
	}
	return errorCode, reason.String(), configs, mSources, hpametrics
}

// RegisterEntry .... mapping input request to elasticserch structure
func RegisterEntry(context *gin.Context) {
	var appRequest models.ApplicationHealthAnalyzeRequest
	//check bad request
	if err := context.BindJSON(&appRequest); err != nil {
		log.Println("Error: encounter context error ", err, " detail ", reflect.TypeOf(err))
		common.ErrorResponse(context, http.StatusBadRequest, "Bad request")
		return
	}
	//check appName
	if common.CheckStrEmpty(appRequest.AppName) {
		log.Println("Error: appName is empty")
		common.ErrorResponse(context, http.StatusBadRequest, "appName is empty")
		return
	}
	//check metric query
	errCode, reason, configs, mSources, hpametrics := convertMetricInfoString(appRequest.Metrics, appRequest.Strategy)
	if errCode != 0 {
		log.Println("encount error while convertMetricInfoString ", reason)
		common.ErrorResponse(context, http.StatusBadRequest, reason)
		return
	}

	var doc models.DocumentRequest

	if appRequest.Strategy == "hpa" {
		doc = models.DocumentRequest{
			appRequest.AppName,
			appRequest.StartTime,
			appRequest.EndTime,
			configs[0],
			configs[1],
			configs[2],
			mSources[0],
			mSources[1],
			mSources[2],
			"200",
			appRequest.Strategy,
			hpametrics,
			appRequest.Policy,
			appRequest.Namespace,
		}
	} else {
		doc = models.DocumentRequest{
			appRequest.AppName,
			appRequest.StartTime,
			appRequest.EndTime,
			configs[0],
			configs[1],
			configs[2],
			mSources[0],
			mSources[1],
			mSources[2],
			"200",
			appRequest.Strategy,
			nil,
			"",
			"",
		}
	}

	log.Printf("DOCREQUEST: %#v\n", doc)

	id, retCode, reason := search.CreateNewDoc(context, elasticClient, doc)
	if retCode == 0 {
		context.JSON(http.StatusOK, converter.ConvertESToNewResp(id, retCode, "new", ""))
	} else {
		common.ErrorResponse(context, http.StatusInternalServerError, reason)
	}
}

// SearchByID .... restful serach by uuid or job id
func SearchByID(context *gin.Context) {
	_id := context.Param("id")
	log.Println("Search by id got called :" + _id + "\n")
	doc, retCode, reason := search.ByID(context, elasticClient, _id)

	if retCode == 0 {
		context.JSON(http.StatusOK, converter.ConvertESToResp(doc))
	} else if retCode == 1 {
		common.ErrorResponse(context, http.StatusNotFound, "failed to retrieve job by id "+_id)
	} else {
		common.ErrorResponse(context, http.StatusInternalServerError, reason)
	}
}

// HpaAlert .... get hpa alert reason
func HpaAlert(context *gin.Context) {
	_appName := context.Param("appName")
	_namespace := context.Param("namespace")
	_strategy := context.Param("strategy")
	id := _appName + ":" + _namespace + ":" + _strategy
	log.Printf("Getting info for %s:%s:%s :\n", _namespace, _appName, _strategy)
	logs, retCode, reason := search.GetLogs(context, elasticClient, id)

	if retCode == 0 {
		context.JSON(http.StatusOK, converter.ConvertESToHPAResp(id, logs))
	} else if retCode == 1 {
		common.ErrorResponse(context, http.StatusNotFound, "failed to retrieve log "+id)
	} else {
		common.ErrorResponse(context, http.StatusInternalServerError, reason)
	}
}

// QueryProxy .... Acting as proxy for different cluster of  query service (for example prometheus)
//                 assume service only access global query service
func QueryProxy(context *gin.Context) {
	//allow crqs
	context.Header("Access-Control-Allow-Origin", "*")
	queryMaps := context.Request.URL.RawQuery
	targeturl := QueryEndpoint + "api/v1/query_range?" + queryMaps
	httpclient := http.Client{
		Timeout: time.Duration(90000 * time.Millisecond),
	}
	resp, err := httpclient.Get(targeturl)
	if err != nil {
		common.ErrorResponse(context, http.StatusBadRequest, "invoke query "+targeturl+" failed ")
		return
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.ErrorResponse(context, http.StatusBadRequest, "failed to retrieve contents "+string(contents))
		return
	}
	context.JSON(http.StatusOK, string(contents))
}

// main .... program entry
func main() {
	var esURL = os.Getenv("ELASTIC_URL")
	QueryEndpoint = os.Getenv("QUERY_SERVICE_ENDPOINT")
	if esURL == "" {
		//esURL = "http://elasticsearch-discovery.foremast.svc.cluster.local:9200/"
		esURL = "http://localhost:9200/"
	}
	if QueryEndpoint == "" {
		// QueryEndpoint = "http://prometheus-k8s.monitoring.svc.cluster.local:9090/"
		QueryEndpoint = "http://a6ac3e9663bb411e9a63702d1928664b-740248454.us-west-2.elb.amazonaws.com:9090/"
	}

	var err error
	// Create Elastic client and wait for Elasticsearch to be ready
	for {
		elasticClient, err = elastic.NewClient(
			elastic.SetURL(esURL),
			elastic.SetSniff(false),
		)
		if err != nil {
			log.Println("failed to reach elasticsearch endpoint ", err)
			// Retry every 3 seconds
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	router := gin.Default()
	v1 := router.Group("/v1/healthcheck")
	{
		//search by id
		v1.GET("/id/:id", SearchByID)
		//create request
		v1.POST("/create", RegisterEntry)

		//v1.POST("/proxyquery", ProxyQuery)
	}
	v2 := router.Group("/api/")
	{
		//query proxy
		v2.GET("/v1/:queryproxy", QueryProxy)
	}

	v3 := router.Group("/alert")
	{
		// get hpa info for namespace/app
		v3.GET("/:appName/:namespace/:strategy", HpaAlert)
	}

	if err = router.Run(":8099"); err != nil {
		log.Fatal(err)
	}
}
