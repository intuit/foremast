package wavefront

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	models "foremast.ai/foremast/foremast-service/pkg/models"
)

// BuildURL .... Build Prometheus url
func BuildURL(metricQuery models.MetricQuery) string {
	config := metricQuery.Parameters
	urlstring := strings.Builder{}
	urlstring.WriteString(url.QueryEscape(config["query"].(string)))
	urlstring.WriteString("&&")
	start := config["start"].(float64)
	urlstring.WriteString(strconv.FormatFloat(start, 'f', 0, 64))
	urlstring.WriteString("&&")
	step := config["step"].(float64)
	if step == 60 {
		urlstring.WriteString("m")
	} else if step == 1 {
		urlstring.WriteString("s")
	} else if step == 3600 {
		urlstring.WriteString("h")
	} else if step == 86400 {
		urlstring.WriteString("d")
	}
	urlstring.WriteString("&&")
	end := config["end"].(float64)
	urlstring.WriteString(strconv.FormatFloat(end, 'f', 0, 64))
	fmt.Println("WAVEFRONT URL: " + urlstring.String())
	return urlstring.String()
}
