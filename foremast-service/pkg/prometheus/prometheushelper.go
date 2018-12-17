package prometheus

import (
	"net/url"
	"strconv"
	"strings"

	models "foremast.ai/foremast/foremast-service/pkg/models"
)

//Build Prometheus url
func BuildUrl(metricQuery models.MetricQuery) string {
	config := metricQuery.Parameters
	urlstring := strings.Builder{}
	urlstring.WriteString(config["endpoint"].(string))
	urlstring.WriteString("query_range?query=")
	urlstring.WriteString(url.QueryEscape(config["query"].(string)))
	urlstring.WriteString("&start=")
	start := config["start"].(float64)
	urlstring.WriteString(strconv.FormatFloat(start, 'f', 0, 64))
	urlstring.WriteString("&end=")
	end := config["end"].(float64)
	urlstring.WriteString(strconv.FormatFloat(end, 'f', 0, 64))
	urlstring.WriteString("&step=")
	urlstring.WriteString(strconv.FormatFloat(config["step"].(float64), 'f', 0, 64))
	return urlstring.String()
}
