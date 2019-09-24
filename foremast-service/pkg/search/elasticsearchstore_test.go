package search

import (
	"foremast.ai/foremast/foremast-service/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockService to return es mock client
func MockService(url string) (*elastic.Client, error) {
	client, err := elastic.NewSimpleClient(elastic.SetURL(url))
	if err != nil {
		return nil, err
	}
	return client, nil
}

var (
	resp = `
		{
			"took": 0,
			"timed_out": false,
			"_shards": {
				"total": 5,
				"successful": 5,
				"skipped": 0,
				"failed": 0
			},
			"hits": {
				"total": 1,
				"max_score": 1.0,
				"hits": [{
					"_index": "documents",
					"_type": "document",
					"_id": "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa",
					"_score": 1.0,
					"_source": {
						"id": "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa",
						"appName": "hpa-samples",
						"created_at": "2019-06-07T00:32:39.079511987Z",
						"startTime": "2019-06-07T00:32:39Z",
						"endTime": "2019-06-07T00:32:39Z",
						"modified_at": "2019-06-08T08:25:04.952641+00:00",
						"currentConfig": "latency== latency",
						"historicalConfig": "latency== latency",
						"currentMetricStore": "latency== prometheus ||traffic== prometheus",
						"historicalMetricStore": "latency== prometheus ||traffic== prometheus",
						"status": "preprocess_inprogress",
						"statusCode": "200",
						"strategy": "hpa",
						"hpaMetricsConfig": {
							"latency": {
								"priority": 2,
								"isIncrease": false,
								"isAbsolute": false
							},
							"traffic": {
								"priority": 1,
								"isIncrease": false,
								"isAbsolute": false
							}
						},
						"namespace": "dev-fm-foremast-examples-usw2-dev-dev",
						"podCountURL": "http://test"
					}
				}]
			}
		}
		`
	handler = http.NotFound
)

// TestByID
func TestByID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	es, _ := MockService(ts.URL)
	doc, retCode, _ := ByID(c, es, "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa")
	assert.NotNil(t, doc)
	assert.Equal(t, int32(0), retCode)
}

// TestGetLogs
func TestGetLogs(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		resp := `
		{
			"took": 0,
			"timed_out": false,
			"_shards": {
				"total": 5,
				"successful": 5,
				"skipped": 0,
				"failed": 0
			},
			"hits": {
				"total": 1,
				"max_score": 1.0,
				"hits": [{
					"_index": "hpalogs",
					"_type": "document",
					"_id": "YoA7LmsBv-DTyYa7RtHz",
					"_score": 1.0,
					"_source": {
						"job_id": "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa",
						"hpalog": {
							"reason": "hpa is scaling down",
							"hpascore": 24,
							"details": [{
								"metricType": "traffic",
								"current": 0.4,
								"upper": 13.28892887803531,
								"lower": 2.347437908396045
							}, {
								"metricType": "latency",
								"current": 2.1917974999998793E-4,
								"upper": 0.6511858437613666,
								"lower": 0.11451766716236265
							}]
						},
						"timestamp": 1.559867604024E9,
						"created_at": "2019-06-06T19:19:25.167323+00:00",
						"modified_at": "2019-06-07T00:33:25.955631+00:00"
					}
				}]
			}
		}
		`
		w.Write([]byte(resp))
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	es, _ := MockService(ts.URL)
	logs, retCode, _ := GetLogs(c, es, "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa")
	assert.Equal(t, 1, len(logs))
	assert.Equal(t, 0, retCode)
}

// TestByStatus
func TestByStatus(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	es, _ := MockService(ts.URL)
	doc, err := ByStatus(es, []string{"preprocess_inprogress", "initial"})
	assert.NotNil(t, doc)
	assert.Nil(t, err)
}

// TestByJobID
func TestByJobID(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	es, _ := MockService(ts.URL)
	uuid, err := ByJobID(c, es, "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa")
	assert.Equal(t, "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa", uuid)
	assert.Nil(t, err)
}

//
func TestCreateNewDoc(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/healthcheck/create", nil)
	es, _ := MockService(ts.URL)
	doc := models.DocumentRequest{
		AppName:            "hpa-samples",
		StartTime:          "2019-06-07T00:32:39Z",
		EndTime:            "2019-06-07T00:32:39Z",
		CurrentConfig:      "latency== latency",
		BaselineConfig:     "baseline_config",
		HistoricalConfig:   "historical_config",
		CurrentMetricStore: "prometheus",
		Namespace:          "test",
		Strategy:           "hpa",
	}
	uuid, _, _ := CreateNewDoc(c, es, doc)
	assert.Equal(t, "hpa-samples:test:hpa", uuid)

	doc2 := models.DocumentRequest{
		AppName:            "hpa-samples",
		StartTime:          "2019-06-07T00:32:39Z",
		EndTime:            "2019-06-07T00:32:39Z",
		CurrentConfig:      "latency== latency",
		BaselineConfig:     "baseline_config",
		HistoricalConfig:   "historical_config",
		CurrentMetricStore: "prometheus",
		Namespace:          "test",
		Strategy:           "rollingupdate",
	}
	uuid2, _, _ := CreateNewDoc(c, es, doc2)
	assert.Equal(t, "c62d7e6ebccb7f6821163a667783b71379b0c26c7192f520b78bc2502190c11b", uuid2)
}

//
func TestConvertDocumentRequestToString(t *testing.T) {
	doc := models.DocumentRequest{
		AppName:            "hpa-samples",
		StartTime:          "2019-06-07T00:32:39Z",
		EndTime:            "2019-06-07T00:32:39Z",
		CurrentConfig:      "latency== latency",
		BaselineConfig:     "baseline_config",
		HistoricalConfig:   "historical_config",
		CurrentMetricStore: "prometheus",
		Namespace:          "test",
		Strategy:           "hpa",
	}
	str := ConvertDocumentRequestToString(doc)
	assert.Equal(t, "hpa-samples2019-06-07T00:32:39Z2019-06-07T00:32:39Zlatency== latencybaseline_confighistorical_configprometheushpa", str)
}
