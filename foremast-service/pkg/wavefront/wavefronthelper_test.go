package wavefront

import (
	"testing"
	"foremast.ai/foremast/foremast-service/pkg/models"
	"github.com/magiconair/properties/assert"
)

// TestBuildURL
func TestBuildURL(t *testing.T) {
	m := models.MetricQuery{}
	m.IsAbsolute = true
	m.IsIncrease = true
	p := 1
	m.Priority = &p
	m.DataSourceType = "wavefront"
	m.Parameters = map[string]interface{}{"query":"test_query", "start": "now-5", "end": "now", "step":1.0}
	str := BuildURL(m)
	assert.Equal(t, str, "test_query&&now-5&&s&&now")
	m.Parameters = map[string]interface{}{"query":"test_query", "start": 5.0, "end": 10.0, "step":60.0}
	str = BuildURL(m)
	assert.Equal(t, str, "test_query&&5&&m&&10")
	m.Parameters = map[string]interface{}{"query":"test_query", "start": 5.0, "end": 10.0, "step":3600.0}
	str = BuildURL(m)
	assert.Equal(t, str, "test_query&&5&&h&&10")
	m.Parameters = map[string]interface{}{"query":"test_query", "start": 5.0, "end": 10.0, "step":86400.0}
	str = BuildURL(m)
	assert.Equal(t, str, "test_query&&5&&d&&10")

}