package wavefront

import (
	"fmt"
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
	urlstring.WriteString(url.QueryEscape(config["query"].(string)))
	urlstring.WriteString("&&")
	var start interface{}
	if reflect.TypeOf(config["start"]).Name() == "string" {
		start = config["start"]
		urlstring.WriteString(start.(string))
	} else {
		start = config["start"].(float64)

		urlstring.WriteString(strconv.FormatFloat(start.(float64), 'f', 0, 64))
	}

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
	var end interface{}
	if reflect.TypeOf(config["start"]).Name() == "string" {
		end = config["end"]
		urlstring.WriteString(end.(string))
	} else {
		end = config["end"].(float64)

		urlstring.WriteString(strconv.FormatFloat(end.(float64), 'f', 0, 64))
	}
	fmt.Println("WAVEFRONT URL: " + urlstring.String())
	return urlstring.String()
}
