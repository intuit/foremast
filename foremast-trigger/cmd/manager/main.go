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
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	fq "foremast.ai/foremast/foremast-trigger/pkg/foremasttrigger"
	// fq "github.com/kianjones4/foremast/foremast-trigger/pkg/foremasttrigger"

	"github.com/gorilla/mux"
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
	Request      fq.ApplicationHealthAnalyzeRequest
}

var jobmap map[string]JobInfo

func CheckJobCompleted(jobID string) int {
	// log.Println(jobID)
	// esUrl := "http://ace26cb17152911e9b3ee067481c81ce-156838986.us-west-2.elb.amazonaws.com:9200/documents/_search"
	esUrl := os.Getenv("FOREMAST_SERVICE_ENDPOINT") + "/v1/healthcheck/" //+ "/documents/_search"
	c, err := fq.NewClient(nil, esUrl)
	healthResponse, err := c.GetStatus(jobID)
	if err != nil {
		log.Printf("getStatus err: %#v\n%s\n", err, err.Error())
		return 1
	}
	// body, _ := ioutil.ReadAll(response.Body)
	log.Printf("healthStatus: %#v\n", healthResponse)
	// healthResponse = fq.ApplicationHealthAnalyzeResponse{}
	// _ = json.Unmarshal(body, &healthResponse)
	if healthResponse.Status == "Healthy" || healthResponse.Status == "Unhealthy" {
		return 0
	} else {
		if healthResponse.Status == "Running" {
			return 2
		} else {
			return 1
		}
	}
}

func CheckAllJobsCompleted() {

}

func StartAnalyzing(analyzingRequest fq.ApplicationHealthAnalyzeRequest) (string, error) {
	//log.Printf("\n\nendpoint: %#v\n\n", analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"])
	endpoint := os.Getenv("FOREMAST_SERVICE_ENDPOINT") + "/v1/healthcheck/create" //os.Getenv("ENDPOINT")

	log.Printf("\nendpoint: %#v\n", endpoint)
	c, err := fq.NewClient(nil, endpoint) //analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"].(string))
	b, err := json.Marshal(analyzingRequest)
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "create"}
	u := c.BaseURL.ResolveReference(rel)

	log.Printf("Request body: %#v\n", string(b))
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {

		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		log.Printf("request err: %#v\n %s\n", err, err.Error())
		return "", err
	}

	// log.Printf("response: %#v\n", resp)

	if resp.StatusCode != 200 {
		return "", errors.NewBadRequest(u.String() + " responded invalid server response:" + resp.Status)
	}
	defer resp.Body.Close()
	var analyzingResponse = fq.ApplicationHealthAnalyzeResponse{}

	//var jobId = ""
	//body, err := ioutil.ReadAll(resp.Body)
	err = json.NewDecoder(resp.Body).Decode(&analyzingResponse)
	if err != nil {
		return "", err
	}

	if analyzingResponse.JobId == "" {
		return "", errors.NewBadRequest(u.String() + " responded invalid server response:" + analyzingResponse.Reason)
	}
	return analyzingResponse.JobId, nil
}

func ForemastQuery(appName string, errorQuery string, latencyQuery string) bool {
	now := time.Now()
	unix := now.Unix()
	startTime := unix - (60 * 5)
	endTime := startTime + (60 * 30)

	analyzingRequest := fq.ApplicationHealthAnalyzeRequest{}

	analyzingRequest.AppName = appName
	fmt.Printf("%s, %s", analyzingRequest.AppName, appName)
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
					"start":    startTime,
					"end":      endTime,
					"step":     60,
				},
			},
			"latency": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    latencyQuery,
					"endpoint": "",
					"start":    startTime,
					"end":      endTime,
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
					"start":    startTime - (7 * 24 * 60 * 60),
					"end":      startTime - (60 * 5),
					"step":     60,
				},
			},
			"latency": {
				DataSourceType: "wavefront",
				Parameters: map[string](interface{}){
					"query":    latencyQuery,
					"endpoint": "",
					"start":    startTime - (7 * 24 * 60 * 60),
					"end":      startTime - (60 * 5),
					"step":     60,
				},
			},
		},
	}

	b, err := json.MarshalIndent(analyzingRequest, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	_ = b
	// log.Printf("%s\n", string(b))

	resp, err := StartAnalyzing(analyzingRequest)
	if err == nil {
		log.Printf("startanalyzing resp %#v\n", resp)
		jobmap[analyzingRequest.AppName] = JobInfo{
			ErrorQuery:   analyzingRequest.Metrics.Current["error5xx"].Parameters["query"].(string),
			LatencyQuery: analyzingRequest.Metrics.Current["latency"].Parameters["query"].(string),
			JobID:        resp,
			Request:      analyzingRequest,
		}
		// log.Printf("jobmap %#v", jobmap)
		return true
	} else {
		log.Printf("startanalyzing err %#v\n%#v\n", err, err.Error())
		return false
	}
	// CheckJobCompleted(resp)

}

func MonitorService(serviceName string) {
	for {
		if CheckJobCompleted(jobmap[serviceName].JobID) <= 1 {
			// create a new job
			ForemastQuery(serviceName, jobmap[serviceName].ErrorQuery, jobmap[serviceName].LatencyQuery)
		} else {
			time.Sleep(time.Second * 10)
		}
	}

}

func serve() {
	router := mux.NewRouter()

	// router.HandleFunc("/restart", ForemastQuery).Methods("GET")

	log.Printf("Service started on port:8011, mode:\n")
	log.Fatal(http.ListenAndServe(":8011", router))

}

func main() {
	jobmap = make(map[string]JobInfo)
	lines := []string{}
	// decoder := json.NewDecoder(r.Body)

	// UpdateTimes(analyzingRequest)
	absPath, _ := filepath.Abs("./requests.csv")
	file, err := os.Open(absPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		// time.Sleep(time.Second * 60)
		// }
	}

	for _, line := range lines {
		values := strings.Split(line, ";")
		success := ForemastQuery(values[0], values[2], values[4])
		for success != true {
			time.Sleep(time.Minute)
			ForemastQuery(values[0], values[2], values[4])
		}
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	for serviceName, _ := range jobmap {
		go MonitorService(serviceName)
	}

	sigTerm := make(chan os.Signal, 1)
	signal.Notify(sigTerm, syscall.SIGTERM)
	signal.Notify(sigTerm, syscall.SIGINT)
	<-sigTerm

}

//func init() {
//	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
//	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
//}
