package main

import (
	common "foremast.ai/foremast/foremast-service/pkg/common"
	converter "foremast.ai/foremast/foremast-service/pkg/converter"
	models "foremast.ai/foremast/foremast-service/pkg/models"
	prometheus "foremast.ai/foremast/foremast-service/pkg/prometheus"
	wavefront "foremast.ai/foremast/foremast-service/pkg/wavefront"
	search "foremast.ai/foremast/foremast-service/pkg/search"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

var (
	elasticClient *elastic.Client
)
// query service endpoint
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

func convertMetricQuerys(metric map[string]models.MetricQuery) (int32, string, string) {
	if len(metric) == 0 {
		return 404, "",""
	}
	output := strings.Builder{}
	metricSourceOutput := strings.Builder{}
	var co int
	for key, value := range metric {
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
	return 0, output.String(),metricSourceOutput.String()
}

func convertMetricInfoString(m models.MetricsInfo, strategy string) (int, string, []string, []string) {

	configs := []string{"", "", ""}
	mSources := []string{"", "", ""}
	if m.Current == nil || len(m.Current) == 0 {
		return 404, "MetricInfo current is empty ", configs,mSources
	}
	errorCode := 0
	reason := strings.Builder{}
	errCode, ret, mSource := convertMetricQuerys(m.Current)

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
		errCode, ret, mSource := convertMetricQuerys(m.Baseline)
		if errCode != 0 {
			log.Println("Warning: baseline convertMetricQuerys ", m.Baseline, " failed. errorCode is ", errCode)
			reason.WriteString(" baseline query encount error ")
			reason.WriteString(ret)
		}
		configs[1] = ret
		mSources[1] = mSource
	}

	if m.Historical != nil {
		hErrCode, ret, mSource := convertMetricQuerys(m.Historical)
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

	return errorCode, reason.String(), configs, mSources
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
	errCode, reason, configs, mSources := convertMetricInfoString(appRequest.Metrics, appRequest.Strategy)
	if errCode != 0 {
		log.Println("encount error while convertMetricInfoString ", reason)
		common.ErrorResponse(context, http.StatusBadRequest, reason)
		return
	}

	doc := models.DocumentRequest{
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
	}
	id, err := search.CreateNewDoc(context, elasticClient, doc)
	context.JSON(http.StatusOK, converter.ConvertESToNewResp(id, err, "new", ""))

}

// SearchByID .... restful serach by uuid or job id
func SearchByID(context *gin.Context) {
	_id := context.Param("id")
	log.Println("Search by id got called :" + _id + "\n")
	doc, err, reason := search.ByID(context, elasticClient, _id)

	if err != 0 {
		if err == -1 {
			context.JSON(http.StatusOK, converter.ConvertESToNewResp(_id, 200, "unknown", _id+" not found."))
		} else {
			context.JSON(http.StatusOK, converter.ConvertESToNewResp(_id, 404, "unknown", reason))
		}
		return
	}
	context.JSON(http.StatusOK, converter.ConvertESToResp(doc))

}

/*
func ProxyQuery(context *gin.Context) {
	var myquery models.QueryRequest
	//check bad request
	if err := context.BindJSON(&myquery); err != nil {
		log.Println("Error: encounter context error ", err, " detail ", reflect.TypeOf(err))
		common.ErrorResponse(context, http.StatusBadRequest, "Bad request")
		return
	}
	httpclient := http.Client{
		Timeout: time.Duration(90000 * time.Millisecond),
	}
	resp, err := httpclient.Get(myquery.QueryString)
	if err != nil {
		common.ErrorResponse(context, http.StatusBadRequest, "invoke query "+myquery.QueryString+" failed " )
		return
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err!= nil{
		common.ErrorResponse(context, http.StatusBadRequest, "failed to retrieve contents "+string(contents))
	}
	context.JSON(http.StatusOK, string(contents))
}*/
// QueryProxy .... Acting as proxy for different cluster of  query service (for example prometheus)
//                 assume service only access global query service
func QueryProxy(context *gin.Context) {
	//allow crqs
	context.Header("Access-Control-Allow-Origin", "*")
	queryMaps := context.Request.URL.RawQuery
	targeturl := QueryEndpoint +"api/v1/query_range?"+queryMaps
	httpclient := http.Client{
		Timeout: time.Duration(90000 * time.Millisecond),
	}
	resp, err := httpclient.Get(targeturl)
	if err != nil {
		common.ErrorResponse(context, http.StatusBadRequest, "invoke query "+targeturl+" failed " )
		return
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err!= nil{
		common.ErrorResponse(context, http.StatusBadRequest, "failed to retrieve contents "+string(contents))
	}
	context.JSON(http.StatusOK, string(contents))
}
// main .... program entry
func main() {
	var esURL = os.Getenv("ELASTIC_URL")
	QueryEndpoint  = os.Getenv("QUERY_SERVICE_ENDPOINT")
	if esURL == "" {
		//esURL = "http://elasticsearch-discovery.foremast.svc.cluster.local:9200/"
		esURL = "http://ace26cb17152911e9b3ee067481c81ce-156838986.us-west-2.elb.amazonaws.com:9200/"
	}
	if QueryEndpoint  ==""{
		QueryEndpoint  = "http://prometheus-k8s.monitoring.svc.cluster.local:9090/"
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
	if err = router.Run(":8099"); err != nil {
		log.Fatal(err)
	}

}
