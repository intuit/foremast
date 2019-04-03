package prometheus

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"

	models "foremast.ai/foremast/foremast-service/pkg/models"
)

// BuildURL .... Build Prometheus url
func BuildURL(metricQuery models.MetricQuery) string {
	config := metricQuery.Parameters
	urlstring := strings.Builder{}
	urlstring.WriteString(config["endpoint"].(string))
	urlstring.WriteString("query_range?query=")
	urlstring.WriteString(url.QueryEscape(config["query"].(string)))
	urlstring.WriteString("&start=")
	var start interface{}
	if reflect.TypeOf(config["start"]).Name() == "string" {
		start = config["start"]
		urlstring.WriteString(start.(string))
	} else {
		start = config["start"].(float64)

		urlstring.WriteString(strconv.FormatFloat(start.(float64), 'f', 0, 64))
	}
	urlstring.WriteString("&end=")
	var end interface{}
	if reflect.TypeOf(config["start"]).Name() == "string" {
		end = config["end"]
		urlstring.WriteString(end.(string))
	} else {
		end = config["end"].(float64)

		urlstring.WriteString(strconv.FormatFloat(end.(float64), 'f', 0, 64))
	}
	urlstring.WriteString("&step=")
	urlstring.WriteString(strconv.FormatFloat(config["step"].(float64), 'f', 0, 64))
	return urlstring.String()
}
