package foremasttrigger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
)

type JobRequest struct {
	Query map[string]map[string][]string `json:"query"`
}

type JobInfo struct {
	JobID        string
	ErrorQuery   string
	LatencyQuery string
	TPSQuery     string
	Request      ApplicationHealthAnalyzeRequest
}

type JobInfoM struct {
	JobID     string
	MetricMap map[string]string
	Request   ApplicationHealthAnalyzeRequest
}

// CheckJobStatus - query the foremast service for the job status
func CheckJobStatus(jobID string, serviceName string) ApplicationHealthAnalyzeResponse {

	esUrl := os.Getenv("FOREMAST_SERVICE_ENDPOINT") + "/v1/healthcheck/"
	c, err := NewClient(nil, esUrl)
	healthResponse, err := c.GetStatus(jobID)
	if err != nil {
		log.Printf("[%s] getStatus err: %#v\n%s\n", serviceName, err, err.Error())
		return ApplicationHealthAnalyzeResponse{Status: "Error"}
	}

	log.Printf("[%s] healthStatus: %#v\n", serviceName, healthResponse)
	return healthResponse
}

// StartAnalyzing - send the request to the foremast service to start the job
func StartAnalyzing(analyzingRequest ApplicationHealthAnalyzeRequest) (string, error) {

	endpoint := os.Getenv("FOREMAST_SERVICE_ENDPOINT") + "/v1/healthcheck/create"
	c, err := NewClient(nil, endpoint)
	if err != nil {
		log.Printf("[%s] error creating client: %#v\n %s\n", analyzingRequest.AppName, err, err.Error())
		return "", err
	}
	b, err := json.Marshal(analyzingRequest)
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "create"}
	u := c.BaseURL.ResolveReference(rel)

	log.Printf("[%s] Request body: %#v\n", analyzingRequest.AppName, string(b))
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {

		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Printf("[%s] request err: %#v\n %s\n", analyzingRequest.AppName, err, err.Error())
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.NewBadRequest(u.String() + " responded invalid server response:" + resp.Status)
	}
	defer resp.Body.Close()
	var analyzingResponse = ApplicationHealthAnalyzeResponse{}

	err = json.NewDecoder(resp.Body).Decode(&analyzingResponse)
	if err != nil {
		return "", err
	}

	if analyzingResponse.JobId == "" {
		return "", errors.NewBadRequest(u.String() + " responded invalid server response:" + analyzingResponse.Reason)
	}
	return analyzingResponse.JobId, nil
}

// QueryWavefrontAnomalyCount - run query for servicename at unix time to get anomaly count from WAVEFRONT
// returns value of anomaly count
func QueryWavefrontAnomalyCount(serviceName string, query string, metricName string, unixtime int64, currentYear *int, currentMonth *time.Month, currentDay *int) float64 {
	client := &http.Client{}

	var Url *url.URL
	Url, err := url.Parse(os.Getenv("WAVEFRONT_ENDPOINT"))
	if err != nil {
		panic("boom")
	}

	Url.Path += "api/v2/chart/api"
	parameters := url.Values{}
	newQuery := strings.Replace(query, "APPNAME", serviceName, -1)
	parameters.Add("q", strings.Replace(newQuery, "REPLACE_METRIC", metricName, -1))
	parameters.Add("s", strconv.FormatInt((unixtime)*1000, 10)) // time in milliseconds
	parameters.Add("g", "d")                                    // get data for past day
	parameters.Add("sorted", "false")
	parameters.Add("cached", "true")
	Url.RawQuery = parameters.Encode()

	log.Printf("[%s] sending request for anomaly count...\n%s", serviceName, Url.String())

	req, err := http.NewRequest("GET", Url.String(), nil)
	if err != nil {
		log.Printf("[%s] error submitting request for anomaly report: %s\n%s\n", serviceName, err, err.Error())
	}
	req.Header.Add("Authorization", "Bearer "+os.Getenv("WAVEFRONT_TOKEN"))
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[%s] error receiving response for anomaly report: %s\n%s\n", serviceName, err, err.Error())
	}
	// log.Printf("%#v\n", resp)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	var responsebody map[string]interface{}
	err = json.Unmarshal(bodyBytes, &responsebody)
	bodyString := string(bodyBytes)
	_ = bodyString
	log.Printf("%#v\n", bodyString)
	if responsebody["warnings"] != nil {
		// no metrics found
		// could mean either no anomalies ever for this metric, or that there was a typo somewhere
		return -1
	}
	// log.Printf("%#v\n", timeseriesdata)
	timeseries := responsebody["timeseries"].([]interface{})
	if len(timeseries) < 1 {
		return 0
	}
	timeseriesdata := responsebody["timeseries"].([]interface{})[0].(map[string]interface{})["data"].([]interface{})[0].([]interface{})[1].(float64)

	return timeseriesdata

}

// generate the anomaly counts for a single service
func GenerateReport(serviceName string, info JobInfoM, currentYear *int, currentMonth *time.Month, currentDay *int) string {
	// custom metric prefix for foremast upper/lower/anomaly metrics
	customMetricPrefix := "custom.iks.foremast."
	// we want a count of anomalies for the metric types over the past day
	baseQuery := `count(ts(REPLACE_METRIC_anomaly, app=APPNAME), app)`
	now := time.Now()
	unix := now.Unix()
	anomalycounts := now.Format("2006-01-02T15:04:05Z07:00")

	for metricName, _ := range info.MetricMap {
		metricAnomalyCount := strconv.FormatFloat(QueryWavefrontAnomalyCount(serviceName, baseQuery, customMetricPrefix+metricName, unix, currentYear, currentMonth, currentDay), 'f', -1, 64)
		anomalycounts = anomalycounts + "\t" + metricAnomalyCount
	}

	return anomalycounts + "\n"
}

// generate the anomaly counts for all services
// func GenerateSummaryReport(serviceslist []string, currentYear *int, currentMonth *time.Month, currentDay *int) {
func GenerateSummaryReport(serviceslist map[string]JobInfoM, currentYear *int, currentMonth *time.Month, currentDay *int) {
	var reports []string

	filename := os.Getenv("VOLUME_PATH") + "/anomalyreport" + strconv.Itoa(*currentYear) + "-" + (*currentMonth).String() + "-" + strconv.Itoa(*currentDay) + ".txt"
	reportfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("error creating file for anomaly report: %s\n%s\n", err, err.Error())
		log.Printf("Failed to generate report")
		return
	}
	defer reportfile.Close()

	// create header for spread sheet
	// currently assumes same metrics being monitored for all services in requests file
	header := "Timestamp\t"
	for _, info := range serviceslist {
		for metric, _ := range info.MetricMap {
			header = header + metric + "\t"
		}
		break
	}
	header = header + "\n"
	reportfile.WriteString(header)

	// generate each of the lines
	for name, info := range serviceslist {
		reports = append(reports, GenerateReport(name, info, currentYear, currentMonth, currentDay))
	}
	for _, report := range reports {
		reportfile.WriteString(report)
	}

	log.Printf("Generated report")
}

// create the initial request to be sent to foremast
func ForemastQuery(jobmap *map[string]JobInfoM, appName string, metricmap map[string]string) bool {
	now := time.Now()
	unix := now.Unix()
	startTime := unix - (60 * 5)
	endTime := startTime + (60 * 30)

	analyzingRequest := ApplicationHealthAnalyzeRequest{}

	analyzingRequest.AppName = appName
	// fmt.Printf("%s, %s", analyzingRequest.AppName, appName)
	analyzingRequest.Strategy = "rollover"
	analyzingRequest.StartTime = now.Format("2006-01-02T15:04:05Z07:00")
	analyzingRequest.EndTime = now.Add(time.Minute * 5).Format("2006-01-02T15:04:05Z07:00") //"2018-11-03T16:33:04-07:00"
	metricsinfo := MetricsInfo{
		Current:    map[string]MetricQuery{},
		Historical: map[string]MetricQuery{},
		Baseline:   map[string]MetricQuery{},
	}
	for name, query := range metricmap {
		mq := MetricQuery{
			DataSourceType: "wavefront",
			Parameters: map[string](interface{}){
				"query":    query,
				"endpoint": "",
				"start":    startTime * 1000,
				"end":      endTime * 1000,
				"step":     60,
			},
		}
		metricsinfo.Current[name] = mq

		mh := MetricQuery{
			DataSourceType: "wavefront",
			Parameters: map[string](interface{}){
				"query":    query,
				"endpoint": "",
				"start":    (startTime - (7 * 24 * 60 * 60)) * 1000,
				"end":      startTime,
				"step":     60,
			},
		}
		metricsinfo.Historical[name] = mh
		metricsinfo.Baseline[name] = mh
	}

	analyzingRequest.Metrics = metricsinfo

	b, err := json.MarshalIndent(analyzingRequest, "", "  ")
	if err != nil {
		fmt.Printf("[%s] error: %s\n", analyzingRequest.AppName, err)
	}
	_ = b

	resp, err := StartAnalyzing(analyzingRequest)
	if err == nil {
		log.Printf("[%s] startanalyzing resp %#v\n", appName, resp)

		(*jobmap)[analyzingRequest.AppName] = JobInfoM{
			MetricMap: metricmap,
			JobID:     resp,
			Request:   analyzingRequest,
		}

		return true
	} else {
		log.Printf("[%s] startanalyzing err %#v\n%#v\n", appName, err, err.Error())
		return false
	}

}

func CreateDashboardURL(serviceName string, jobmap *map[string]JobInfoM, healthresponse ApplicationHealthAnalyzeResponse) string {
	// the custom prefix for al of the upper/lower bound metrics
	customMetricPrefix := "custom.iks.foremast."
	dashboardUrl := ""

	metricRegex := `&quot;name&quot;\s*:\s*&quot;([\w\.]*)`
	r := regexp.MustCompile(metricRegex)
	matches := r.FindStringSubmatch(healthresponse.Reason)
	if len(matches) < 2 {
		log.Printf("No metric in health response: %s", healthresponse.Reason)
		dashboardUrl = os.Getenv("WAVEFRONT_ENDPOINT") + "/dashboard/Foremast"
	} else {
		// base wavefront url that contains metrics for the app metric, foremast upper bound, lower bound and anomaly data
		dashboardUrl = os.Getenv("WAVEFRONT_ENDPOINT") + `/chart#_v01(c:(cs:(type:line),id:chart,n:%22REPLACE_CUSTOM_METRIC%22,s:!((co:'rgb(247,12,28)',e:'',n:Query,q:'avg(ts(REPLACE_CUSTOM_METRIC_upper,%20app=%22$%7Bapp_name%7D%22),%20app)',qb:'%7B%22_v%22:1,%22metric%22:%22ts(REPLACE_CUSTOM_METRIC_upper)%22,%22filters%22:%5B%5B%5B%22app%22,%22$%7Bapp_name%7D%22%5D%5D,%22and%22%5D,%22functions%22:%5B%5B%22avg%22,%5B%22app%22%5D%5D%5D%7D',qbe:!t,s:Y),(co:'rgb(0,0,255)',e:'',n:'New%20Query',q:'avg(ts(REPLACE_CUSTOM_METRIC_lower,%20app=%22$%7Bapp_name%7D%22),%20app)',qb:'%7B%22_v%22:1,%22metric%22:%22ts(REPLACE_CUSTOM_METRIC_lower)%22,%22filters%22:%5B%5B%5B%22app%22,%22$%7Bapp_name%7D%22%5D%5D,%22and%22%5D,%22functions%22:%5B%5B%22avg%22,%5B%22app%22%5D%5D%5D%7D',qbe:!t,s:Y),(co:'rgb(0,246,47)',e:'',n:'New%20Query',q:'avg(ts(REPLACE_CUSTOM_METRIC_anomaly,%20app=%22$%7Bapp_name%7D%22),%20app)',qb:'%7B%22_v%22:1,%22metric%22:%22ts(REPLACE_CUSTOM_METRIC_anomaly)%22,%22filters%22:%5B%5B%5B%22app%22,%22$%7Bapp_name%7D%22%5D%5D,%22and%22%5D,%22functions%22:%5B%5B%22avg%22,%5B%22app%22%5D%5D%5D%7D',qbe:!t,s:Y),(co:'rgba(185,0,255,1)',e:'',n:'New%20Query',q:'REPLACE_QUERY',qbe:!f,s:Y))),g:(c:off,d:7200,ls:!t,s:REPLACE_TIME,w:'2h'),p:(app_name:REPLACE_APP))`
		dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_CUSTOM_METRIC", customMetricPrefix+strings.ToLower(matches[1]), -1)
		dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_QUERY", (*jobmap)[serviceName].MetricMap[strings.ToLower(matches[1])], -1)

	}

	timeRegex := `&quot;ts&quot;\s*:\s*\[(\d*).\d`

	r = regexp.MustCompile(timeRegex)
	matches = r.FindStringSubmatch(healthresponse.Reason)

	if len(matches) < 2 {
		log.Printf("No timestamp in health response: %s", healthresponse.Reason)
		dashboardUrl = os.Getenv("WAVEFRONT_ENDPOINT") + "/dashboard/Foremast"
	} else {
		timestamp := strings.ToLower(matches[1])
		newtime, _ := strconv.ParseInt(timestamp, 0, 64)
		newtime = newtime - (60 * 15)

		dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_APP", serviceName, -1)
		dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_TIME", strconv.FormatInt(newtime, 10), -1)
	}

	return dashboardUrl
}

// continuously monitors a service by polling
func MonitorService(serviceName string, mutex *sync.Mutex, anomalyfilename *string, jobmap *map[string]JobInfoM, currentYear *int, currentMonth *time.Month, currentDay *int) {
	for {
		// check if the job completed
		healthresponse := CheckJobStatus((*jobmap)[serviceName].JobID, serviceName)
		status := healthresponse.Status
		if status == "Healthy" {
			// we're done with this job, run the next query
			ForemastQuery(jobmap, serviceName, (*jobmap)[serviceName].MetricMap)
		} else if status == "Unhealthy" {
			//we're done with this job, write the anomaly to file and run next query

			dashboardUrl := CreateDashboardURL(serviceName, jobmap, healthresponse)

			log.Printf("[%s] dashboardUrl: %s", serviceName, dashboardUrl)

			s := time.Now().Format("2006-01-02T15:04:05Z07:00") + "\t" + serviceName + "\t" + (*jobmap)[serviceName].JobID + "\t" + healthresponse.Reason + "\t" + dashboardUrl + "\n" //timestamp + servicename + jobid
			mutex.Lock()

			// create daily anoamly files, so ensure that nothing's changed since the last update
			cury, curm, curd := time.Now().Date()
			if cury != *currentYear || curm != *currentMonth || curd != *currentDay {
				*currentYear = cury
				*currentMonth = curm
				*currentDay = curd
				*anomalyfilename = os.Getenv("VOLUME_PATH") + "/anomaly_" + strconv.Itoa(*currentYear) + "-" + (*currentMonth).String() + "-" + strconv.Itoa(*currentDay) + ".tsv"
			}

			// open and update file
			anomalyPath, _ := filepath.Abs(*anomalyfilename)
			anomalyfile, err := os.OpenFile(anomalyPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
			if err != nil {
				log.Printf("[%s] error opening anomaly file %s\n%s\n", serviceName, err, err.Error())
			}
			_, err = anomalyfile.WriteString(html.UnescapeString(s))
			if err != nil {
				log.Printf("[%s] error writing to anomaly file %s\n%s\n", serviceName, err, err.Error())
			}
			anomalyfile.Close()
			mutex.Unlock()

			ForemastQuery(jobmap, serviceName, (*jobmap)[serviceName].MetricMap)
		} else if status == "Abort" || status == "Warning" {
			// give up and run another query
			ForemastQuery(jobmap, serviceName, (*jobmap)[serviceName].MetricMap)
		} else {
			// not done, check again later
			time.Sleep(time.Second * 10)
		}
	}

}
