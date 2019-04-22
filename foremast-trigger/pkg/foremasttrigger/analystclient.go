package foremasttrigger

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
)

type Interface interface {
	StartAnalyzing(namespace string, appName string, podNames [][]string, endpoint string, metrics Metrics, timeWindow time.Duration) (string, error)
	GetStatus(jobId string) (ApplicationHealthAnalyzeResponse, error)
}

type Client struct {
	BaseURL   *url.URL
	UserAgent string

	HttpClient *http.Client
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
	Metrics MetricsInfo `json:"metrics"`

	// canary or blue-green
	Strategy string `json:"strategy"`
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
}

func NewClient(HttpClient *http.Client, endpoint string) (*Client, error) {
	if HttpClient == nil {
		HttpClient = http.DefaultClient
	}
	if endpoint != "" {
		u, err := url.Parse(endpoint)
		if err != nil {
			log.Printf("The base URL is wrong: %v", endpoint)
			return nil, err
		} else {
			return &Client{HttpClient: HttpClient, BaseURL: u}, nil
		}
	} else {
		return &Client{HttpClient: HttpClient}, nil
	}
}

//const TIME_FORMAT = "2006-01-02T15:04:05-07:00"

//func (c *Client) toMetricQueries()

func (c *Client) StartAnalyzing(namespace string, appName string, podNames [][]string, endpoint string, metrics Metrics, timeWindow time.Duration, strategy string) (string, error) {
	//queries[] MetricQuery

	var t = time.Now()
	var startTime = t.Format(time.RFC3339)
	t = t.Add(timeWindow * time.Minute)
	var endTime = t.Format(time.RFC3339)

	var metricsInfo, err = CreateMetricsInfo(namespace, appName, podNames, metrics, timeWindow, strategy)
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
	}

	b, err := json.Marshal(analyzingRequest)
	if err != nil {
		return "", err
	}

	rel := &url.URL{Path: "create"}
	u := c.BaseURL.ResolveReference(rel)

	log.Printf("Request body: %v", string(b))
	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
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

/*
[
    {
        "id": "5f99629670a33e99ecf6a7218b8351f6fedeba9fc4284dd5968fa02d298ccc80",
        "appName": "k8s-metrics-demo",
        "created_at": "2018-10-26T00:50:03.457664Z",
        "startTime": "2018-10-26T00:50:03.457665Z",
        "endTime": "2018-10-26T00:50:03.457665Z",
        "modified_at": "2018-10-26T00:50:03.457665Z",
        "status": "created",
        "stage": "initial",
        "content": "start=\"\ufffd\",end=\"\ufffd\",endpoint=\"http://localhost:9090/api/v1/query_range\",filterStr=\"namespace_pod:http_server_requests_error_4xx\",steps=\"60\""
    }
]
*/

func (c *Client) GetStatus(jobId string) (ApplicationHealthAnalyzeResponse, error) {
	//queries[] MetricQuery
	rel := &url.URL{Path: "id/" + jobId}
	u := c.BaseURL.ResolveReference(rel)
	// log.Printf("statusrequesturl %#v\n", u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return ApplicationHealthAnalyzeResponse{
			StatusCode: 500,
		}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return ApplicationHealthAnalyzeResponse{
			StatusCode: 501,
		}, err
	}
	defer resp.Body.Close()

	//var response []map[string]string
	// log.Printf("StatusResponse %#v\n", resp)

	var response = ApplicationHealthAnalyzeResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return ApplicationHealthAnalyzeResponse{
			StatusCode: 503,
		}, err
	}

	/*
			switch status {
		case "initial":
			return "new"
		case "preprocess_inprogress":
			return "inprogress"
		case "postprocess_inprogress":
			return "inprogress"
		case "completed_health":
			return "success"
		case "completed_unhealth":
			return "anomaly"
		case "abort":
			return "abort"
		default:
			return "unknown"
		}
	*/
	var phase string
	switch response.Status {
	case "created", "initial", "new", "inprogress":
		phase = MonitorPhaseRunning
		break
	case "completed_health", "success":
		phase = MonitorPhaseHealthy
		break
	case "completed_unhealth", "anomaly":
		phase = MonitorPhaseUnhealthy
		break
	case "abort", "unknown":
		phase = MonitorPhaseAbort
		break
	case "completed_unknown":
		phase = MonitorPhaseWarning
		break
	default:
		phase = response.Status
	}

	//var getStatusResponse = ApplicationHealthAnalyzeResponse{
	//	StatusCode: 200,
	//	Status:     phase,
	//}
	response.Status = phase
	// log.Printf("StatusResponse %#v\n", response)
	return response, nil
}
