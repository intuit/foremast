package main

import (
	"net/http/httptest"
	"testing"

	"bufio"
	"foremast.ai/foremast/foremast-service/pkg/common"
	"foremast.ai/foremast/foremast-service/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"time"
	)

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
						"historicalMetricStore": "latency== prometheus ||traffic== wavefront",
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
		}`
)

// TestGetOpenRequest ...
func TestGetOpenRequest(t *testing.T) {
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
	client, _ := elastic.NewSimpleClient(elastic.SetURL(ts.URL))
	elasticClient = client
	GetOpenRequest(c)
	queue := common.GetQueueInstance()
	res := models.DocumentResponse{"test_id", "test", time.Now(), time.Now(),
		time.Now(), time.Now().Add(time.Duration(-50) * time.Second), "hpa", "", "", "",
		"prometheus", "prometheus", "prometheus", "initial", "200", "",
		"", "", nil, "", "test", "current", ""}
	queue.Push(res)
	GetOpenRequest(c)
	queue.Push(res)
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{something_wrong}`))
	}
	GetOpenRequest(c)
}

// TestFillQueueFromES ...
func TestFillQueueFromES(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	client, _ := elastic.NewSimpleClient(elastic.SetURL(ts.URL))
	elasticClient = client
	fillQueueFromES()
	queue := common.GetQueueInstance()
	assert.Equal(t, 0, queue.Len())
	time.Sleep(time.Duration(2 * time.Second))
	assert.Equal(t, 1, queue.Len())
}

// TestHpaAlert ...
func TestHpaAlert(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	client, _ := elastic.NewSimpleClient(elastic.SetURL(ts.URL))
	elasticClient = client
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	HpaAlert(c)
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"took": 0,
			"timed_out": false,
			"_shards": {
				"total": 5,
				"successful": 5,
				"skipped": 0,
				"failed": 0
			},
			"hits": {
				"total": 0,
				"max_score": 1.0,
				"hits": []
			}
		}`))
	}
	HpaAlert(c)
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`"{hits": {
				"total": 0,
				"max_score": 1.0,
				"hits": []
			}
		}`))
	}
	HpaAlert(c)
}

// TestSearchByID ...
func TestSearchByID(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	client, _ := elastic.NewSimpleClient(elastic.SetURL(ts.URL))
	elasticClient = client
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	SearchByID(c)
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"took": 0,
			"timed_out": false,
			"_shards": {
				"total": 5,
				"successful": 5,
				"skipped": 0,
				"failed": 0
			},
			"hits": {
				"total": 0,
				"max_score": 1.0,
				"hits": []
			}
		}`))
	}
	SearchByID(c)
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`"{hits": {
				"total": 0,
				"max_score": 1.0,
				"hits": []
			}
		}`))
	}
	SearchByID(c)
}

// TestRegisterEntry ...
func TestRegisterEntry(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	handler = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(resp))
	}
	client, _ := elastic.NewSimpleClient(elastic.SetURL(ts.URL))
	elasticClient = client
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/Register", bufio.NewReader(strings.NewReader(`{"AppName": "test"}`)))
	RegisterEntry(c)
	c.Request = httptest.NewRequest("GET", "/Register", bufio.NewReader(strings.NewReader(`test`)))
	RegisterEntry(c)
	c.Request = httptest.NewRequest("GET", "/Register", bufio.NewReader(strings.NewReader(`{}`)))
	RegisterEntry(c)

	content := `
		{
			"appName": "hpa-samples2",
			"startTime": "2019-04-23T18:11:16Z",
			"endTime": "2019-04-23T18:21:16Z",
			"metrics": {
				"current": {
					"cpu": {
						"dataSourceType": "prometheus",
						"priority": 2,
						"parameters": {
							"end": 1556043720,
							"endpoint": "http://localhost:9090/api/v1/",
							"query": "namespace_app_pod_cpu_usage_seconds_total{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
							"start": 1556043120,
							"step": 60
						}
					},
					"tomcat_threads": {
						"dataSourceType": "prometheus",
						"priority": 3,
						"parameters": {
							"end": 1556043720,
							"endpoint": "http://localhost:9090/api/v1/",
							"query": "namespace_app_pod_tomcat_threads_busy_percentage{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
							"start": 1556043120,
							"step": 60
						}
					},
					"traffic": {
						"dataSourceType": "prometheus",
						"priority": 2,
						"parameters": {
							"end": 1556043720,
							"endpoint": "http://localhost:9090/api/v1/",
							"query": "namespace_app_pod_http_server_requests_count{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
							"start": 1556043120,
							"step": 60
						}
					}
				},
				"historical": {
					"cpu": {
						"dataSourceType": "prometheus",
						"priority": 3,
						"parameters": {
							"end": 1556043060,
							"endpoint": "http://localhost:9090/api/v1/",
							"query": "namespace_app_pod_cpu_usage_seconds_total{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
							"start": 1555438260,
							"step": 60
						}
					},
					"tomcat_threads": {
						"dataSourceType": "prometheus",
						"priority": 3,
						"parameters": {
							"end": 1556043060,
							"endpoint": "http://localhost:9090/api/v1/",
							"query": "namespace_app_pod_tomcat_threads_busy_percentage{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
							"start": 1555438260,
							"step": 60
						}
					},
					"traffic": {
						"dataSourceType": "prometheus",
						"priority": 4,
						"parameters": {
							"end": 1556043060,
							"endpoint": "http://localhost:9090/api/v1/",
							"query": "namespace_app_pod_http_server_requests_count{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
							"start": 1555438260,
							"step": 60
						}
					}
				}
			},
			"podCountURL": {

				"dataSourceType": "prometheus",

				"parameters": {
					"end": 1556043720,
					"endpoint": "http://localhost:9090/api/v1/",
					"query": "namespace_app_pod_count{namespace=\"dev-fm-foremast-examples-usw2-dev-dev\",app=\"hpa-samples\"}",
					"start": 1556043120,
					"step": 60
				}

			},
			"strategy": "hpa",
			"namespace": "dev-fm-foremast-examples-usw2-dev-dev"
		}
	`
	c.Request = httptest.NewRequest("GET", "/Register", bufio.NewReader(strings.NewReader(content)))
	RegisterEntry(c)
	c.Request = httptest.NewRequest("GET", "/Register", bufio.NewReader(strings.NewReader(strings.Replace(content, "hpa", "nohpa", -1))))
	RegisterEntry(c)
}

// TestOther to cover other code
func TestOther(t *testing.T) {
	priority := 1
	q := models.MetricQuery{DataSourceType: "prometheus", Priority: &priority, IsIncrease: true, IsAbsolute: true}
	v, _, _ := constructURL(q)
	assert.Equal(t, int32(404), v)
	q.Parameters = map[string]interface{}{"endpoint": "http://localhost/", "query":"test_query", "start": "now-5", "end": "now", "step":60.0}
	_, _, s2 := constructURL(q)
	assert.Equal(t, "http://localhost/query_range?query=test_query&start=now-5&end=now&step=60", s2)
	q.DataSourceType = "wavefront"
	_, _, s2 = constructURL(q)
	assert.Equal(t, "test_query&&now-5&&m&&now", s2)
	q.DataSourceType = "unknown"
	_, s, _ := constructURL(q)
	assert.Equal(t, "unknown", s)
	code, _, _ := convertMetricQuerys(nil, "")
	assert.Equal(t, int32(404), code)
	code, _, _ = convertMetricQuerys(map[string]models.MetricQuery{"": q}, "hpa")
	assert.Equal(t, int32(404), code)
	info := models.MetricsInfo{ Current: map[string]models.MetricQuery{"": q}, Baseline: map[string]models.MetricQuery{"": q}, Historical: map[string]models.MetricQuery{"": q}}
	c, _, _, _, _ := convertMetricInfoString(info, "")
	assert.Equal(t, 404, c)
	info.Historical = nil
	c, _, _, _, _ = convertMetricInfoString(info, "")
	assert.Equal(t, 404, c)
}
