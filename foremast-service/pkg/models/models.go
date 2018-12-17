package models

import "time"

//MetricQuery
type MetricQuery struct {
	DataSourceType string `json:"dataSourceType"`

	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Can be any metric data Source Type
	// For prometheus dataSource Type
	//Name string `json:"name,omitempty"`
	//// For example: error4xx error5xx cpu memory latency etc.

	//Endpoint string `json:"endpoint,omitempty"`
	//
	//Query string `json:"query,omitempty"`
	//
	//Step int32 `json:"step,omitempty"`
	//
	//Start int64 `json:"start,omitempty"`
	//
	//End int64 `json:"end,omitempty"`
}

type MetricsInfo struct {
	Current    map[string]MetricQuery `json:"current"`
	Baseline   map[string]MetricQuery `json:"baseline,omitempty"`
	Historical map[string]MetricQuery `json:"historical,omitempty"`
}

type ApplicationHealthAnalyzeRequest struct {
	AppName string `json:"appName"`

	//RFC3339     = "2006-01-02T15:04:05+07:00"
	StartTime string `json:"startTime"`

	//RFC3339     = "2006-01-02T15:04:05+07:00"
	EndTime string `json:"endTime"`

	// error4xx error5xx cpu memory latency
	//MetricNames []string `json:"metricNames"`

	//// key: current, baseline, historical
	Metrics MetricsInfo `json:"metrics"`

	// canary or blue-green
	Strategy string `json:"strategy"`
}

type AnomalyInfo struct {
	Tags   string  `json:"tags"`
	Values []int64 `json:"values"`
}

type ApplicationHealthAnalyzeResponse struct {
	JobId      string                 `json:"jobId"`
	StatusCode int32                  `json:"statusCode"`
	Status     string                 `json:"status"`
	Reason     string                 `json:"reason,omitempty"`
	Anomaly    map[string]AnomalyInfo `json:"anomaly",omitempty`
}

type ApplicationHealthAnalyzeResponseNew struct {
	JobId      string `json:"jobId"`
	StatusCode int32  `json:"statusCode"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
}

type Document struct {
	ID                string    `json:"id"`
	AppName           string    `json:"appName"`
	CreatedAt         time.Time `json:"created_at"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
	ModifiedAt        time.Time `json:"modified_at"`
	CurrentConfig     string    `json:"currentConfig"`
	BaselineConfig    string    `json:"baselineConfig",omitempty`
	HistoricalConfig  string    `json:"historicalConfig",omitempty`
	Status            string    `json:"status"`
	StatusCode        string    `json:"statusCode"`
	Strategy          string    `json:"strategy"`
	Reason            string    `json:"reason",omitempty`
	ProcessingContent string    `json:"processingContent",omitempty`
}

type DocumentRequest struct {
	AppName          string `json:"appName"`
	StartTime        string `json:"startTime"`
	EndTime          string `json:"endTime"`
	CurrentConfig    string `json:"contentConfig"`
	BaselineConfig   string `json:"baselineConfig",omitempty`
	HistoricalConfig string `json:"historicalConfig",omitempty`
	StatusCode       string `json:"statusCode",omitempty`
	Strategy         string `json:"strategy"`
}

type DocumentResponse struct {
	ID         string    `json:"id"`
	AppName    string    `json:"appName"`
	CreatedAt  time.Time `json:"created_at"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	ModifiedAt time.Time `json:"modified_at"`
	//Type       string    `json:"type"`
	Strategy          string `json:"strategy"`
	CurrentConfig     string `json:"currentConfig"`
	BaselineConfig    string `json:"baselineConfig",omitempty`
	HistoricConfig    string `json:"historicConfig",omitempty`
	Status            string `json:"status"`
	StatusCode        string `json:"statusCode"`
	Reason            string `json:"reason",omitempty`
	ProcessingContent string `json:"processingContent",omitempty`
	AnomalyInfo       string `json:"anomalyInfo",omitempty`
}

type SearchResponse struct {
	Time      string             `json:"time"`
	Hits      string             `json:"hits"`
	Documents []DocumentResponse `json:"documents"`
}
