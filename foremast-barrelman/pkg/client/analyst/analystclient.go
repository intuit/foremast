package analyst

import (
	"bytes"
	"encoding/json"
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	m "foremast.ai/foremast/foremast-barrelman/pkg/client/metrics"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/glog"
	"net/http"
	"net/url"
	"time"
)

type Interface interface {
	StartAnalyzing(namespace string, appName string, podNames [][]string, endpoint string, metrics d.Metrics, timeWindow time.Duration) (string, error)
	GetStatus(jobId string) (ApplicationHealthAnalyzeResponse, error)
}

type Client struct {
	BaseURL   *url.URL
	UserAgent string

	DoFunc func(req *http.Request) (*http.Response, error)
}

type ApplicationHealthAnalyzeRequest struct {
	AppName string `json:"appName"`

	//RFC3339     = "2006-01-02T15:04:05Z07:00"
	StartTime string `json:"startTime"`

	//RFC3339     = "2006-01-02T15:04:05Z07:00"
	EndTime string `json:"endTime"`

	// error4xx error5xx cpu memory latency
	//MetricNames []string `json:"metricNames"`

	//// key: current, baseline, historical
	Metrics m.MetricsInfo `json:"metrics"`

	// canary or blue-green
	Strategy string `json:"strategy"`

	// Namespace
	Namespace string `json:"namespace,omitempty"`

	// Pod count url
	PodCountURL m.MetricQuery `json:"podCountURL,omitempty"`
}

type AnomalyInfo struct {
	Tags   string    `json:"tags"`
	Values []float64 `json:"values"`
}

type ApplicationHealthAnalyzeResponse struct {
	StatusCode int32 `json:"statusCode"`

	Reason string `json:"reason,omitempty"`

	JobId string `json:"jobId"`

	Status string `json:"status"`

	Anomaly map[string]AnomalyInfo `json:"anomaly,omitempty"`

	HpaLogs []d.HpaLogEntry `json:"hpaLogs"`
}

/**
HPA history
{
    "job_id": "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa",
    "hpalogs": [{
            "hpalog": {
                "details": [{
                        "current": 2.7000000000000006,
                        "lower": 0,
                        "metricAlias": "traffic",
                        "upper": 1.4194502009551655
                    },
                    {
                        "current": 4,
                        "lower": 0,
                        "metricAlias": "tomcat_threads",
                        "upper": 1.7711655929134411
                    },
                    {
                        "current": 0.040515150723457634,
                        "lower": -0.0005548594374170275,
                        "metricAlias": "cpu",
                        "upper": 0.011646879268297602
                    }
                ],
                "hpascore": 55,
                "reason": "hpa is scaling up"
            },
            "timestamp": "0001-01-01T00:00:00Z"
        }

    ]
}
*/

func NewClient(httpClient *http.Client, endpoint string) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if endpoint != "" {
		u, err := url.Parse(endpoint)
		if err != nil {
			glog.V(5).Infof("The base URL is wrong: %v", endpoint)
			return nil, err
		} else {
			return &Client{DoFunc: httpClient.Do, BaseURL: u}, nil
		}
	} else {
		return &Client{DoFunc: httpClient.Do}, nil
	}
}

//const TIME_FORMAT = "2006-01-02T15:04:05-07:00"

//func (c *Client) toMetricQueries()

func (c *Client) StartAnalyzing(namespace string, appName string, podNames [][]string, endpoint string, metrics d.Metrics,
	timeWindow time.Duration, strategy string, metricAliases []string) (string, error) {
	//queries[] MetricQuery

	var t = time.Now()
	var startTime = t.Format(time.RFC3339)
	t = t.Add(timeWindow * time.Minute)
	var endTime = t.Format(time.RFC3339)

	var metricsInfo, err = m.CreateMetricsInfo(namespace, appName, podNames, metrics, timeWindow, strategy, metricAliases)
	if err != nil {
		return "", err
	}
	//basic_date_time_no_millis
	//A basic formatter that combines a basic date and time without millis, separated by a T: yyyyMMdd'T'HHmmssZ.
	var analyzingRequest = ApplicationHealthAnalyzeRequest{
		AppName:   appName,
		StartTime: startTime,
		EndTime:   endTime,
		Strategy:  strategy,
		Metrics:   metricsInfo,
		Namespace: namespace,
	}
	podCountURL, err := m.CreatePodCountURL(namespace, appName, metrics, timeWindow)
	if err == nil {
		analyzingRequest.PodCountURL = podCountURL
	}

	b, err := json.Marshal(analyzingRequest)
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "create"}
	u := c.BaseURL.ResolveReference(rel)

	glog.Infof("Request body: %v", string(b))
	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.DoFunc(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.NewBadRequest(u.String() + " responded invalid server response:" + resp.Status)
	}
	defer resp.Body.Close()
	var analyzingResponse = ApplicationHealthAnalyzeResponse{}

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

func (c *Client) GetStatus(jobId string) (ApplicationHealthAnalyzeResponse, error) {
	//queries[] MetricQuery
	rel := &url.URL{Path: "id/" + jobId}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return ApplicationHealthAnalyzeResponse{
			StatusCode: 500,
		}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.DoFunc(req)
	if err != nil {
		return ApplicationHealthAnalyzeResponse{
			StatusCode: 501,
		}, err
	}
	defer resp.Body.Close()

	//var response []map[string]string

	var response = ApplicationHealthAnalyzeResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ApplicationHealthAnalyzeResponse{
			StatusCode: 503,
		}, err
	}

	var phase string
	switch response.Status {
	case "created", "initial", "new", "inprogress", "unknown":
		phase = d.MonitorPhaseRunning
		break
	case "completed_health", "success":
		phase = d.MonitorPhaseHealthy
		break
	case "completed_unhealth", "anomaly":
		phase = d.MonitorPhaseUnhealthy
		break
	case "abort":
		phase = d.MonitorPhaseAbort
		break
	case "completed_unknown":
		phase = d.MonitorPhaseWarning
		break
	default:
		phase = response.Status
	}

	response.Status = phase
	return response, nil
}
