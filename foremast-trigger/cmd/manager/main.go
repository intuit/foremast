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
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	fq "foremast.ai/foremast/foremast-trigger/pkg/foremasttrigger"
	// fq "github.com/kianjones4/foremast/foremast-trigger/pkg/foremasttrigger"
)

var (
	masterURL  string
	kubeconfig string
)

var jobmap map[string]fq.JobInfo
var serviceslist []string
var anomalyfilename string = "/data/anomaly/anomaly.tsv"
var currentYear int
var currentMonth time.Month
var currentDay int

func main() {
	jobmap = make(map[string]fq.JobInfo)
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

	fq.GenerateSummaryReport(serviceslist, &currentYear, &currentMonth, &currentDay)

	for _, line := range lines {
		//appname;error4xx;errorquery;latency;latencyquery;tps;tpsquery
		values := strings.Split(line, ";")
		success := fq.ForemastQuery(values[0], values[2], values[4], values[6], &jobmap)
		for success != true {
			time.Sleep(time.Second * 10)
			success = fq.ForemastQuery(values[0], values[2], values[4], values[6], &jobmap)
		}
	}

	stopCh := make(chan struct{})
	defer close(stopCh)

	mutex := &sync.Mutex{}
	_ = mutex

	anomalyfilename = "/data/anomaly/anomaly_" + strconv.Itoa(currentYear) + "-" + currentMonth.String() + "-" + strconv.Itoa(currentDay) + ".tsv"

	for serviceName, _ := range jobmap {
		go fq.MonitorService(serviceName, mutex, &anomalyfilename, &jobmap, &currentYear, &currentMonth, &currentDay)

	}

	go func() {
		time.Sleep(time.Hour * 24)
		fq.GenerateSummaryReport(serviceslist, &currentYear, &currentMonth, &currentDay)
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
