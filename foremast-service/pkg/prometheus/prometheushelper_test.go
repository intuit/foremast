package prometheus

import (
	"testing"
	"foremast.ai/foremast/foremast-service/pkg/models"
	"github.com/magiconair/properties/assert"
	)

// TestbuildURL
func TestBuildURL(t *testing.T) {
	m := models.MetricQuery{}
	m.IsAbsolute = true
	m.IsIncrease = true
	p := 1
	m.Priority = &p
	m.DataSourceType = "prometheus"
	m.Parameters = map[string]interface{}{"endpoint": "http://localhost/", "query":"test_query", "start": "now-5", "end": "now", "step":60.0}
	str := BuildURL(m)
	assert.Equal(t, str, "http://localhost/query_range?query=test_query&start=now-5&end=now&step=60")
	m.Parameters = map[string]interface{}{"endpoint": "http://localhost/", "query":"test_query", "start": 10.0, "end": 15.0, "step":60.0}
	str2 := BuildURL(m)
	assert.Equal(t, str2, "http://localhost/query_range?query=test_query&start=10&end=15&step=60")
}
