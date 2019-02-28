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

// type JobQuery {
//   Terms JobTerms
// }
//
// type JobTerms {
//
// }

func CheckJobCompleted(jobID string) bool {
	esUrl := "http://ace26cb17152911e9b3ee067481c81ce-156838986.us-west-2.elb.amazonaws.com:9200/documents/_search"
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
		} else {
			time.Sleep(time.Second * 10)
		}

	}

	return false
}

func StartAnalyzing(analyzingRequest fq.ApplicationHealthAnalyzeRequest) (string, error) {
	//log.Printf("\n\nendpoint: %#v\n\n", analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"])
	log.Printf("\nendpoint: %#v\n", os.Getenv("ENDPOINT"))
	c, err := fq.NewClient(nil, os.Getenv("ENDPOINT")) //analyzingRequest.Metrics.Current["error4xx"].Parameters["endpoint"].(string))
	b, err := json.Marshal(analyzingRequest)
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "create"}
	u := c.BaseURL.ResolveReference(rel)

	// log.Printf("Request body: %v\n", string(b))
	req, err := http.NewRequest("POST", os.Getenv("ENDPOINT"), bytes.NewReader(b))
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

func ForemastQuery(w http.ResponseWriter, r *http.Request) {
	// decoder := json.NewDecoder(r.Body)

	analyzingRequest := fq.ApplicationHealthAnalyzeRequest{}
	// err := decoder.Decode(&analyzingRequest)
	// if err != nil {
	// 	panic(err)
	// }
	// log.Printf("%#v", analyzingRequest)
	absPath, _ := filepath.Abs("./requestsinput.txt")
	// absPath, _ := filepath.Abs("siminput/simrequests0.txt")
	file, err := os.Open(absPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		byt := []byte(scanner.Text())
		if err := json.Unmarshal(byt, &analyzingRequest); err != nil {
			panic(err)
		}
		// log.Printf("%#v\n", analyzingRequest)
		resp, err := StartAnalyzing(analyzingRequest)
		if err == nil {
			log.Printf("startanalyzing resp %#v\n", resp)
		}
		CheckJobCompleted(resp)
	}

}

func serve() {
	router := mux.NewRouter()

	// router.HandleFunc("/restart", ForemastQuery).Methods("GET")

	log.Printf("Service started on port:8011, mode:\n")
	log.Fatal(http.ListenAndServe(":8011", router))

}

func main() {
	//serve()
	// for true {
	ForemastQuery(nil, nil)
	// time.Sleep(time.Second * 60)
	// }

}

//func init() {
//	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
//	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
//}
