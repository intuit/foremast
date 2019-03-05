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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	fq "foremast.ai/foremast/foremast-query/pkg/foremastquery"
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

func CheckJobCompleted(jobID string) bool {
	// esUrl := "http://ace26cb17152911e9b3ee067481c81ce-156838986.us-west-2.elb.amazonaws.com:9200/documents/_search"
	esUrl := "http://a2c1d2f06186b11e98f4602f39e94cef-36929393.us-west-2.elb.amazonaws.com:9200/documents/_search"
	newRequest := JobRequest{
		Query: map[string]map[string][]string{
			"terms": {
				"_id": []string{jobID},
			},
		},
	}

	requestbody, err := json.Marshal(newRequest)
	if err != nil {
		log.Printf("req error %s\n", err)
		return false
	}
	log.Printf("reqbody: %#v", string(requestbody))

	completed := false
	c, err := fq.NewClient(nil, os.Getenv("ENDPOINT"))

	for completed != true {
		req, err := http.NewRequest("GET", esUrl, bytes.NewReader(requestbody))
		if err != nil {
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		resp, err := c.HttpClient.Do(req)
		if err != nil {
			log.Printf("request err: %#v", err)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		// log.Printf("respbody: %#v\n", string(body))
		respmap := map[string]interface{}{}
		_ = json.Unmarshal(body, &respmap)
		// log.Printf("%#v\n", respmap["hits"])
		if len(respmap["hits"].(map[string]interface{})["hits"].([]interface{})) < 1 {
			continue
		}
		status := respmap["hits"].(map[string]interface{})["hits"].([]interface{})[0].(map[string]interface{})["_source"].(map[string]interface{})["status"].(string)
		log.Printf("%s status %#v\n", jobID[0:9], status)
		if strings.HasPrefix(status, "completed") {
			completed = true
			return true
		} else {
			time.Sleep(time.Second * 10)
		}

	}

	return false
}

func CheckAllJobsCompleted() {

}

func StartAnalyzing(analyzingRequest fq.ApplicationHealthAnalyzeRequest) (string, error) {
	//log.Printf("\n\nendpoint: %#v\n\n", analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"])
	endpoint := "http://" + os.Getenv("FOREMAST_SERVICE_SERVICE_HOST") + ":" + os.Getenv("FOREMAST_SERVICE_SERVICE_PORT_HTTP") + "/v1/healthcheck/create" //os.Getenv("ENDPOINT")

	log.Printf("\nendpoint: %#v\n", endpoint)
	c, err := fq.NewClient(nil, endpoint) //analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"].(string))
	b, err := json.Marshal(analyzingRequest)
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "create"}
	u := c.BaseURL.ResolveReference(rel)

	// log.Printf("Request body: %v\n", string(b))
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
	nanos := now.UnixNano()
	millis := nanos / 1000000
	startTime := millis - (60 * 5)
	endTime := startTime + (60 * 30)

	sample := `{"appName":"fds-dac","startTime":"2018-11-03T16:50:04-07:00","endTime":"2018-11-03T16:33:04-07:00","strategy":"rollover","metrics":{"current":{"error4xx":{"dataSourceType":"wavefront","parameters":{"end":1550825400000,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"","start":1550815200000,"step":60}},"latency":{"dataSourceType":"wavefront","parameters":{"end":1550825400000,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"avg(align(60s, mean, ts(appdynamics.apm.transactions.avg_resp_time_ms, bu=cto and app=fds-dac and env=prd and source=fds-dac )), name)","start":1550815200000,"step":60}}},"historical":{"error4xx":{"dataSourceType":"wavefront","parameters":{"end":1550825400000,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"avg(align(60s, mean, ts(appdynamics.apm.transactions.errors_per_min, bu=cto and app=fds-dac and env=prd and source=fds-dac )), name)","start":1550815200000,"step":60}},"latency":{"dataSourceType":"wavefront","parameters":{"end":1550825400000,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"avg(align(60s, mean, ts(appdynamics.apm.transactions.avg_resp_time_ms, bu=cto and app=fds-dac and env=prd and source=fds-dac )), name)","start":1550815200000,"step":60}}},"strategy":"rollover"}}`

	analyzingRequest := fq.ApplicationHealthAnalyzeRequest{}
	byt := []byte(sample)
	if err := json.Unmarshal(byt, &analyzingRequest); err != nil {
		panic(err)
	}

	endpoint := "http://" + os.Getenv("FOREMAST_SERVICE_SERVICE_HOST") + ":" + os.Getenv("FOREMAST_SERVICE_SERVICE_PORT_HTTP") + "/v1/healthcheck/create" //os.Getenv("ENDPOINT")

	analyzingRequest.AppName = appName
	fmt.Printf("%s, %s", analyzingRequest.AppName, appName)
	analyzingRequest.Metrics.Current["error4xx"].Parameters["query"] = errorQuery
	analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"] = endpoint
	analyzingRequest.Metrics.Current["error4xx"].Parameters["start"] = startTime
	analyzingRequest.Metrics.Current["error4xx"].Parameters["end"] = endTime
	analyzingRequest.Metrics.Current["latency"].Parameters["query"] = latencyQuery
	analyzingRequest.Metrics.Current["latency"].Parameters["endpoint"] = endpoint
	analyzingRequest.Metrics.Current["latency"].Parameters["start"] = startTime
	analyzingRequest.Metrics.Current["latency"].Parameters["end"] = endTime
	analyzingRequest.Metrics.Historical["error4xx"].Parameters["query"] = errorQuery
	analyzingRequest.Metrics.Historical["error4xx"].Parameters["endpoint"] = endpoint
	analyzingRequest.Metrics.Historical["error4xx"].Parameters["start"] = startTime
	analyzingRequest.Metrics.Historical["error4xx"].Parameters["end"] = endTime
	analyzingRequest.Metrics.Historical["latency"].Parameters["query"] = latencyQuery
	analyzingRequest.Metrics.Historical["latency"].Parameters["endpoint"] = endpoint
	analyzingRequest.Metrics.Historical["latency"].Parameters["start"] = startTime
	analyzingRequest.Metrics.Historical["latency"].Parameters["end"] = endTime
	b, err := json.MarshalIndent(analyzingRequest, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	_ = b
	// fmt.Print(string(b))

	resp, err := StartAnalyzing(analyzingRequest)
	if err == nil {
		// log.Printf("startanalyzing resp %#v\n", resp)
		jobmap[analyzingRequest.AppName] = JobInfo{
			ErrorQuery:   analyzingRequest.Metrics.Current["error4xx"].Parameters["query"].(string),
			LatencyQuery: analyzingRequest.Metrics.Current["latency"].Parameters["query"].(string),
			JobID:        resp,
			Request:      analyzingRequest,
		}
		log.Printf("jobmap %#v", jobmap)
		return true
	} else {
		log.Printf("startanalyzing err %#v\n%#v\n", err, err.Error())
		return false
	}
	// CheckJobCompleted(resp)

}

func serve() {
	router := mux.NewRouter()

	// router.HandleFunc("/restart", ForemastQuery).Methods("GET")

	log.Printf("Service started on port:8011, mode:\n")
	log.Fatal(http.ListenAndServe(":8011", router))

}

func main() {
	jobmap = make(map[string]JobInfo)

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
		values := strings.Split(line, ";")
		success := ForemastQuery(values[0], values[2], values[4])
		for success != true {
			time.Sleep(time.Second * 3)
			ForemastQuery(values[0], values[2], values[4])
		}
		// time.Sleep(time.Second * 60)
		// }
	}

	for {
		if CheckJobCompleted(jobmap["fds-fpp"].JobID) {
			// UpdateTimes(&jobmap[k].Request)
			ForemastQuery("fds-fpp", jobmap["fds-fpp"].ErrorQuery, jobmap["fds-fpp"].LatencyQuery)
		}
	}

}

//func init() {
//	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
//	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
//}
