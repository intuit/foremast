/*
Copyright 2018 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	fq "foremast.ai/foremast/foremast-trigger/pkg/foremasttrigger"
	// fq "github.com/kianjones4/foremast/foremast-trigger/pkg/foremasttrigger"

	"k8s.io/apimachinery/pkg/api/errors"
)

var (
	masterURL  string
	kubeconfig string
)

type JobRequest struct {
	Query map[string]map[string][]string `json:"query"`
}

type JobInfo struct {
	JobID        string
	ErrorQuery   string
	LatencyQuery string
	TPSQuery     string
	Request      fq.ApplicationHealthAnalyzeRequest
}

var jobmap map[string]JobInfo
var serviceslist []string
var anomalyfilename string = "./anomaly.tsv"
var currentYear int
var currentMonth time.Month
var currentDay int

// CheckJobCompleted - query the foremast service for the job status
func CheckJobCompleted(jobID string, serviceName string) fq.ApplicationHealthAnalyzeResponse {

	esUrl := os.Getenv("FOREMAST_SERVICE_ENDPOINT") + "/v1/healthcheck/"
	c, err := fq.NewClient(nil, esUrl)
	healthResponse, err := c.GetStatus(jobID)
	if err != nil {
		log.Printf("[%s] getStatus err: %#v\n%s\n", serviceName, err, err.Error())
		return fq.ApplicationHealthAnalyzeResponse{Status: "Error"}
	}

	log.Printf("[%s] healthStatus: %#v\n", serviceName, healthResponse)
	return healthResponse
}

// StartAnalyzing - send the request to the foremast service to start the job
func StartAnalyzing(analyzingRequest fq.ApplicationHealthAnalyzeRequest) (string, error) {

	endpoint := os.Getenv("FOREMAST_SERVICE_ENDPOINT") + "/v1/healthcheck/create"
	c, err := fq.NewClient(nil, endpoint)
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
	var analyzingResponse = fq.ApplicationHealthAnalyzeResponse{}

	err = json.NewDecoder(resp.Body).Decode(&analyzingResponse)
	if err != nil {
		return "", err
	}

	if analyzingResponse.JobId == "" {
		return "", errors.NewBadRequest(u.String() + " responded invalid server response:" + analyzingResponse.Reason)
	}
	return analyzingResponse.JobId, nil
}

// QueryWavefront - run query for servicename at unix time to get anomaly count from WAVEFRONT
// returns value of anomaly count
func QueryWavefront(serviceName string, query string, unixtime int64) float64 {
	client := &http.Client{}

	var Url *url.URL
	Url, err := url.Parse(os.Getenv("WAVEFRONT_ENDPOINT"))
	if err != nil {
		panic("boom")
	}

	Url.Path += "api/v2/chart/api"
	parameters := url.Values{}
	parameters.Add("q", strings.Replace(query, "APPNAME", serviceName, -1))
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
	req.Header.Add("Authorization", os.Getenv("WAVEFRONT_TOKEN"))
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
func GenerateReport(serviceName string) string {
	baseTpsQuery := `count(ts("custom.iks.foremast.appdynamics.apm.transactions.calls_per_min_anomaly", app=APPNAME), app)`
	baseErrorQuery := `count(ts("custom.iks.foremast.appdynamics.apm.transactions.errors_per_min_anomaly", app=APPNAME), app)`
	baseLatencyQuery := `count(ts("custom.iks.foremast.appdynamics.apm.transactions.90th_percentile_resp_time_ms_anomaly", app=APPNAME), app)`

	now := time.Now()
	unix := now.Unix()

	tpsAnomalyCount := strconv.FormatFloat(QueryWavefront(serviceName, baseTpsQuery, unix), 'f', -1, 64)
	errorAnomalyCount := strconv.FormatFloat(QueryWavefront(serviceName, baseErrorQuery, unix), 'f', -1, 64)
	latencyAnomalyCount := strconv.FormatFloat(QueryWavefront(serviceName, baseLatencyQuery, unix), 'f', -1, 64)

	return now.Format("2006-01-02T15:04:05Z07:00") + "\t" + serviceName + "\t" + tpsAnomalyCount + "\t" + errorAnomalyCount + "\t" + latencyAnomalyCount + "\n"

}

// generate the anomaly counts for all services
func GenerateSummaryReport() {
	var reports []string

	// now := time.Now()
	// unix := now.Unix()
	// filename := "./anomalyreport" + strconv.FormatInt(unix*1000, 10) + ".txt"
	filename := "./anomalyreport" + strconv.Itoa(currentYear) + "-" + currentMonth.String() + "-" + strconv.Itoa(currentDay) + ".txt"
	reportfile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		log.Printf("error creating file for anomaly report: %s\n%s\n", err, err.Error())
	}
	defer reportfile.Close()
	reportfile.WriteString("Timestamp\tService\tTPS\tError\tLatency\n")
	for _, name := range serviceslist {
		reports = append(reports, GenerateReport(name))
	}
	for _, report := range reports {
		reportfile.WriteString(report)
	}

	log.Printf("Generated report")
	// time.Sleep(time.Hour * 24)
}

// create the initial request to be sent to foremast
func ForemastQuery(appName string, errorQuery string, latencyQuery string, tpsQuery string) bool {
	now := time.Now()
	unix := now.Unix()
	startTime := unix - (60 * 5)
	endTime := startTime + (60 * 30)

	analyzingRequest := fq.ApplicationHealthAnalyzeRequest{}

	analyzingRequest.AppName = appName
	// fmt.Printf("%s, %s", analyzingRequest.AppName, appName)
	analyzingRequest.Strategy = "rollover"
	analyzingRequest.StartTime = now.Format("2006-01-02T15:04:05Z07:00")
	analyzingRequest.EndTime = now.Add(time.Minute * 5).Format("2006-01-02T15:04:05Z07:00") //"2018-11-03T16:33:04-07:00"
	analyzingRequest.Metrics = fq.MetricsInfo{
		Current: map[string]fq.MetricQuery{
			"error5xx": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    errorQuery,
					"endpoint": "",
					"start":    startTime * 1000,
					"end":      endTime * 1000,
					"step":     60,
				},
			},
			"latency": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    latencyQuery,
					"endpoint": "",
					"start":    startTime * 1000,
					"end":      endTime * 1000,
					"step":     60,
				},
			},
			"tps": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    tpsQuery,
					"endpoint": "",
					"start":    startTime * 1000,
					"end":      endTime * 1000,
					"step":     60,
				},
			},
		},
		Historical: map[string]fq.MetricQuery{
			"error5xx": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    errorQuery,
					"endpoint": "",
					"start":    (startTime - (7 * 24 * 60 * 60)) * 1000,
					"end":      startTime,
					"step":     60,
				},
			},
			"latency": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    latencyQuery,
					"endpoint": "",
					"start":    (startTime - (7 * 24 * 60 * 60)) * 1000,
					"end":      startTime,
					"step":     60,
				},
			},
			"tps": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    tpsQuery,
					"endpoint": "",
					"start":    (startTime - (7 * 24 * 60 * 60)) * 1000,
					"end":      endTime,
					"step":     60,
				},
			},
		},
	}

	b, err := json.MarshalIndent(analyzingRequest, "", "  ")
	if err != nil {
		fmt.Printf("[%s] error: %s\n", analyzingRequest.AppName, err)
	}
	_ = b
	// log.Printf("%s\n", string(b))

	resp, err := StartAnalyzing(analyzingRequest)
	if err == nil {
		log.Printf("[%s] startanalyzing resp %#v\n", appName, resp)
		jobmap[analyzingRequest.AppName] = JobInfo{
			ErrorQuery:   analyzingRequest.Metrics.Current["error5xx"].Parameters["query"].(string),
			LatencyQuery: analyzingRequest.Metrics.Current["latency"].Parameters["query"].(string),
			TPSQuery:     analyzingRequest.Metrics.Current["tps"].Parameters["query"].(string),
			JobID:        resp,
			Request:      analyzingRequest,
		}
		// log.Printf("jobmap %#v", jobmap)
		return true
	} else {
		log.Printf("[%s] startanalyzing err %#v\n%#v\n", appName, err, err.Error())
		return false
	}
	// CheckJobCompleted(resp)

}

// continuously monitors a service by polling
func MonitorService(serviceName string, mutex *sync.Mutex, filename string) {
	for {
		healthresponse := CheckJobCompleted(jobmap[serviceName].JobID, serviceName)
		status := healthresponse.Status
		if status == "Healthy" {
			// we're done with this job, run the next query
			ForemastQuery(serviceName, jobmap[serviceName].ErrorQuery, jobmap[serviceName].LatencyQuery, jobmap[serviceName].TPSQuery)
		} else if status == "Unhealthy" {
			//we're done with this job, write the anomaly to file and run next query
			timeRegex := `&quot;ts&quot;\s*:\s*\[(\d*).\d`

			r := regexp.MustCompile(timeRegex)
			matches := r.FindStringSubmatch(healthresponse.Reason)

			// dashboardUrl := os.Getenv("WAVEFRONT_ENDPOINT") + `/dashboard/Foremast#_v01(g:(c:off,d:7200,ls:!f,s:REPLACE_TIME,w:'2h'),p:(app:(d:Label,f:TAG_KEY,k:app,l:Application,m:(Label_2:REPLACE_APP),q:'ts(appdynamics.apm.errors.errors_per_min,%20app=%22fds-*%22)',s:Label_2)))`
			dashboardUrl := os.Getenv("WAVEFRONT_ENDPOINT") + `/chart#_v01(c:(cs:(type:line),id:chart,n:%22REPLACE_METRIC_lower%22,s:!((co:'rgb(247,12,28)',e:'',n:Query,q:'avg(ts(%22REPLACE_METRIC_upper%22,%20app=%22REPLACE_APP%22),%20app)',qb:'%7B%22_v%22:1,%22metric%22:%22ts(%5C%22REPLACE_METRIC_upper%5C%22)%22,%22filters%22:%5B%5B%5B%22app%22,%22REPLACE_APP%22%5D%5D,%22and%22%5D,%22functions%22:%5B%5B%22avg%22,%5B%22app%22%5D%5D%5D%7D',qbe:!t,s:Y),(co:'rgb(0,0,0)',e:'',n:'New%20Query',q:'avg(ts(%22REPLACE_METRIC_lower%22,%20app=%22REPLACE_APP%22),%20app)',qb:'%7B%22_v%22:1,%22metric%22:%22ts(%5C%22REPLACE_METRIC_lower%5C%22)%22,%22filters%22:%5B%5B%5B%22app%22,%22REPLACE_APP%22%5D%5D,%22and%22%5D,%22functions%22:%5B%5B%22avg%22,%5B%22app%22%5D%5D%5D%7D',qbe:!t,s:Y),(co:'rgb(166,7,231)',e:'',n:'New%20Query',q:'avg(ts(REPLACE_WF_METRIC,%20app=%22REPLACE_APP%22),%20app)',qb:'%7B%22_v%22:1,%22metric%22:%22ts(REPLACE_WF_METRIC)%22,%22filters%22:%5B%5B%5B%22app%22,%22REPLACE_APP%22%5D%5D,%22and%22%5D,%22functions%22:%5B%5B%22avg%22,%5B%22app%22%5D%5D%5D%7D',qbe:!t,s:Y))),g:(c:off,d:7200,ls:!f,s:1553641280,w:'2h'))`
			if len(matches) < 2 {
				log.Printf("No timestamp in health response: %s", healthresponse.Reason)
				dashboardUrl = os.Getenv("WAVEFRONT_ENDPOINT") + "/dashboard/Foremast"
			} else {
				timestamp := strings.ToLower(matches[1])

				dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_APP", serviceName, -1)
				dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_TIME", timestamp, -1)
			}

			metricRegex := `&quot;name&quot;\s*:\s*&quot;([\w\.]*)`
			r = regexp.MustCompile(metricRegex)
			matches = r.FindStringSubmatch(healthresponse.Reason)
			if len(matches) < 2 {
				log.Printf("No metric in health response: %s", healthresponse.Reason)
				dashboardUrl = os.Getenv("WAVEFRONT_ENDPOINT") + "/dashboard/Foremast"
			} else {

				dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_METRIC", strings.ToLower(matches[1]), -1)
				if matches[1] == "custom.iks.foremast.appdynamics.apm.transactions.errors_per_min" {
					dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_WF_METRIC", "appdynamics.apm.errors.errors_per_min", -1)
				} else if matches[1] == "custom.iks.foremast.appdynamics.apm.transactions.errors_per_min" {
					dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_WF_METRIC", "appdynamics.apm.transactions.90th_percentile_resp_time_ms", -1)
				} else {
					dashboardUrl = strings.Replace(dashboardUrl, "REPLACE_WF_METRIC", "appdynamics.apm.transactions.calls_per_min", -1)
				}

			}

			s := time.Now().Format("2006-01-02T15:04:05Z07:00") + "\t" + serviceName + "\t" + jobmap[serviceName].JobID + "\t" + healthresponse.Reason + "\t" + dashboardUrl + "\n" //timestamp + servicename + jobid
			mutex.Lock()

			cury, curm, curd := time.Now().Date()
			if cury != currentYear || curm != currentMonth || curd != currentDay {
				currentYear = cury
				currentMonth = curm
				currentDay = curd
				anomalyfilename = "./anomaly_" + strconv.Itoa(currentYear) + "-" + currentMonth.String() + "-" + strconv.Itoa(currentDay) + ".tsv"
			}

			anomalyPath, _ := filepath.Abs(anomalyfilename)
			anomalyfile, err := os.OpenFile(anomalyPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
			if err != nil {
				log.Printf("error opening anomaly file %s\n%s\n", err, err.Error())
			}
			_, err = anomalyfile.WriteString(html.UnescapeString(s))
			if err != nil {
				log.Printf("error writing to anomaly file %s\n%s\n", err, err.Error())
			}
			anomalyfile.Close()
			mutex.Unlock()
			ForemastQuery(serviceName, jobmap[serviceName].ErrorQuery, jobmap[serviceName].LatencyQuery, jobmap[serviceName].TPSQuery)
		} else if status == "Abort" || status == "Warning" {
			// give up and run another query
			ForemastQuery(serviceName, jobmap[serviceName].ErrorQuery, jobmap[serviceName].LatencyQuery, jobmap[serviceName].TPSQuery)
		} else {
			// not done, check again later
			time.Sleep(time.Second * 10)
		}
	}

}

func main() {
	jobmap = make(map[string]JobInfo)
	lines := []string{}

	// decoder := json.NewDecoder(r.Body)

	// UpdateTimes(analyzingRequest)
	requestsfilename := os.Getenv("REQUESTS_FILE")
	// requestsPath, _ := filepath.Abs("./requests.csv")
	requestsPath, _ := filepath.Abs("./" + requestsfilename)
	requestsfile, err := os.Open(requestsPath)
	if err != nil {
		log.Fatal(err)
	}
	defer requestsfile.Close()

	scanner := bufio.NewScanner(requestsfile)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		// time.Sleep(time.Second * 60)
		// }
	}

	for _, line := range lines {
		values := strings.Split(line, ";")
		serviceslist = append(serviceslist, values[0])
	}

	now := time.Now()
	currentYear, currentMonth, currentDay = now.Date()

	GenerateSummaryReport()

	for _, line := range lines {
		values := strings.Split(line, ";")
		success := ForemastQuery(values[0], values[2], values[4], values[6])
		for success != true {
			time.Sleep(time.Second * 10)
			success = ForemastQuery(values[0], values[2], values[4], values[6])
		}
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	mutex := &sync.Mutex{}
	_ = mutex

	anomalyfilename = "./anomaly_" + strconv.Itoa(currentYear) + "-" + currentMonth.String() + "-" + strconv.Itoa(currentDay) + ".tsv"

	for serviceName, _ := range jobmap {
		go MonitorService(serviceName, mutex, anomalyfilename)

	}

	go func() {
		time.Sleep(time.Hour * 24)
		GenerateSummaryReport()
	}()

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm

}

//func init() {
//	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
//	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
//}
