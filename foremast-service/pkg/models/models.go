package models

import "time"

// MetricQuery .... define metric query structures
type MetricQuery struct {
	DataSourceType string `json:"dataSourceType"`

	Parameters map[string]interface{} `json:"parameters,omitempty"`

	Priority *int `json:"priority,omitempty"`

	IsIncrease bool `json:"isIncrease,omitempty"`

	IsAbsolute bool `json:"isAbsolute,omitempty"`

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

// MetricsInfo .... MetricsInfo structure by current, baseline and historical
type MetricsInfo struct {
	Current    map[string]MetricQuery `json:"current"`
	Baseline   map[string]MetricQuery `json:"baseline,omitempty"`
	Historical map[string]MetricQuery `json:"historical,omitempty"`
}

// ApplicationHealthAnalyzeRequest .... structure by appname, start, end time metric info and strategy
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

	// list of metrics and their priorities for HPA
	HPAMetrics []HPAMetric `json:"hpaMetrics,omitempty"`

	// for later
	Policy string `json:"policy,omitempty"`

	// app namespace
	Namespace string `json:"namespace,omitempty"`

	// pod count metric URL
	PodCountURL MetricQuery `json:"podCountURL,omitempty"`
}

// QueryRequest qurey string as struct
type QueryRequest struct {
	QueryString string `json:"queryString"`
}

// AnomalyInfo .... anomaly structures
type AnomalyInfo struct {
	Tags   string  `json:"tags"`
	Values []int64 `json:"values"`
}

// ApplicationHealthAnalyzeResponse  -- health analyze response fields
type ApplicationHealthAnalyzeResponse struct {
	JobID      string                   `json:"jobId"`
	StatusCode int32                    `json:"statusCode"`
	Status     string                   `json:"status"`
	Reason     string                   `json:"reason,omitempty"`
	Anomaly    map[string]AnomalyInfo   `json:"anomaly,omitempty"`
	HPALog     []map[string]interface{} `json:"hpalogs,omitempty"`
}

// ApplicationHealthAnalyzeResponseNew .... new response
type ApplicationHealthAnalyzeResponseNew struct {
	JobID      string `json:"jobId"`
	StatusCode int32  `json:"statusCode"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
}

// Document --- elastic search document index
type Document struct {
	ID                    string               `json:"id"`
	AppName               string               `json:"appName"`
	CreatedAt             time.Time            `json:"created_at"`
	StartTime             time.Time            `json:"startTime"`
	EndTime               time.Time            `json:"endTime"`
	ModifiedAt            time.Time            `json:"modified_at"`
	CurrentConfig         string               `json:"currentConfig"`
	BaselineConfig        string               `json:"baselineConfig,omitempty"`
	HistoricalConfig      string               `json:"historicalConfig,omitempty"`
	CurrentMetricStore    string               `json:"currentMetricStore,omitempty"`
	BaselineMetricStore   string               `json:"baselineMetricStore,omitempty"`
	HistoricalMetricStore string               `json:"historicalMetricStore,omitempty"`
	Status                string               `json:"status"`
	StatusCode            string               `json:"statusCode"`
	Strategy              string               `json:"strategy"`
	Reason                string               `json:"reason,omitempty"`
	ProcessingContent     string               `json:"processingContent,omitempty"`
	HPAMetrics            map[string]HPAMetric `json:"hpaMetricsConfig,omitempty"`
	Policy                string               `json:"policy,omitempty"`
	Namespace             string               `json:"namespace,omitempty"`
	PodCountURL           string               `json:"podCountURL,omitempty"`
}

// DocumentRequest .... request structure
type DocumentRequest struct {
	AppName               string               `json:"appName"`
	StartTime             string               `json:"startTime"`
	EndTime               string               `json:"endTime"`
	CurrentConfig         string               `json:"currentConfig"`
	BaselineConfig        string               `json:"baselineConfig,omitempty"`
	HistoricalConfig      string               `json:"historicalConfig,omitempty"`
	CurrentMetricStore    string               `json:"currentMetricStore,omitempty"`
	BaselineMetricStore   string               `json:"baselineMetricStore,omitempty"`
	HistoricalMetricStore string               `json:"historicalMetricStore,omitempty"`
	StatusCode            string               `json:"statusCode,omitempty"`
	Strategy              string               `json:"strategy"`
	HPAMetrics            map[string]HPAMetric `json:"hpaMetricsConfig,omitempty"`
	Policy                string               `json:"policy,omitempty"`
	Namespace             string               `json:"namespace,omitempty"`
	PodCountURL           string               `json:"podCountURL,omitempty"`
}

// DocumentResponse .... es response structure
type DocumentResponse struct {
	ID         string    `json:"id"`
	AppName    string    `json:"appName"`
	CreatedAt  time.Time `json:"created_at"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
	ModifiedAt time.Time `json:"modified_at"`
	//Type       string    `json:"type"`
	Strategy            string               `json:"strategy"`
	CurrentConfig       string               `json:"currentConfig"`
	BaselineConfig      string               `json:"baselineConfig,omitempty"`
	HistoricConfig      string               `json:"historicConfig,omitempty"`
	CurrentMetricStore  string               `json:"currentMetricStore,omitempty"`
	BaselineMetricStore string               `json:"baselineMetricStore,omitempty"`
	HistoricMetricStore string               `json:"historicMetricStore,omitempty"`
	Status              string               `json:"status"`
	StatusCode          string               `json:"statusCode"`
	Reason              string               `json:"reason,omitempty"`
	ProcessingContent   string               `json:"processingContent,omitempty"`
	AnomalyInfo         string               `json:"anomalyInfo,omitempty"`
	HPAMetrics          map[string]HPAMetric `json:"hpaMetricsConfig,omitempty"`
	Policy              string               `json:"policy,omitempty"`
	Namespace           string               `json:"namespace,omitempty"`
}

// SearchResponse .... es search reqponse structure
type SearchResponse struct {
	Time      string             `json:"time"`
	Hits      string             `json:"hits"`
	Documents []DocumentResponse `json:"documents"`
}

// HPAMetric as detail of hpa metric
type HPAMetric struct {
	Priority   int  `json:"priority"`
	IsIncrease bool `json:"isIncrease"`
	IsAbsolute bool `json:"isAbsolute"`
}

// HPALogResponse combine logs to single entity with job id
type HPALogResponse struct {
	JobID      string   `json:"jobId"`
	HPALog     []HPALog `json:"hpalogs"`
	StatusCode int32    `json:"statusCode"`
	Reason     string   `json:"reason,omitempty"`
}

// HPALog hpa log by foremast-brain to know detail of score change
type HPALog struct {
	JobID      string     `json:"job_id,omitempty"`
	ModifiedAt *time.Time `json:"modified_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	Timestamp  float64    `json:"timestamp"`
	Log        struct {
		HPAScore int    `json:"hpascore"`
		Reason   string `json:"reason"`
		Details  []struct {
			MetricType string  `json:"metricType"`
			Current    float64 `json:"current"`
			Upper      float64 `json:"upper"`
			Lower      float64 `json:"lower"`
		} `json:"details"`
	} `json:"hpalog"`
}
