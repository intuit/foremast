package analyst

import (
	"bytes"
	"fmt"
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func normalResponse(req *http.Request) (*http.Response, error) {
	// just in case you want default correct return value
	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(normalRepsonse))),
	}, nil
}

var normalRepsonse = `
{
    "statusCode": 200,
    "status": "success",
    "jobId": "hpa-samples:dev-fm-foremast-examples-usw2-dev-dev:hpa",
    "hpalogs": [{
            "hpalog": {
                "details": [{
                        "current": 2.7000000000000006,
                        "lower": 0,
                        "metricAlias": "traffic",
                        "upper": 1.4194502009551655
                    },
                    {
                        "current": 4,
                        "lower": 0,
                        "metricAlias": "tomcat_threads",
                        "upper": 1.7711655929134411
                    },
                    {
                        "current": 0.040515150723457634,
                        "lower": -0.0005548594374170275,
                        "metricAlias": "cpu",
                        "upper": 0.011646879268297602
                    }
                ],
                "hpascore": 55,
                "reason": "hpa is scaling up"
            },
            "timestamp": "0001-01-01T00:00:00Z"
        }

    ]
}
`

func TestPayload(*testing.T) {

	var u, _ = url.Parse("http://127.0.0.1:8099/api/test")
	var client = &Client{DoFunc: normalResponse, BaseURL: u}

	var timeWindow = time.Duration(30 * 1000)
	var metricAliases = make([]string, 2)
	metricAliases[0] = "cpu"
	metricAliases[1] = "memory"

	var monitoring = make([]d.Monitoring, 2)
	monitoring[0] = d.Monitoring{
		MetricName:  "namespace_app_pod_cpu_utilization",
		MetricAlias: "cpu",
		MetricType:  "guage",
	}
	monitoring[1] = d.Monitoring{
		MetricName:  "namespace_app_pod_jvm_heap_utilization",
		MetricAlias: "memory",
		MetricType:  "guage",
	}

	jobId, err := client.StartAnalyzing("default", "k8s-metrics-demo", nil, "", d.Metrics{
		DataSourceType: "prometheus",
		Endpoint:       "",
		Monitoring:     monitoring,
	}, timeWindow, "hpa",
		metricAliases)
	if err != nil {
		glog.Error(err)
		return
	}

	fmt.Println(jobId)

	getStatusResponse, err := client.GetStatus(jobId)
	if err != nil {
		glog.Error(err)
		return
	}

	fmt.Println(getStatusResponse.Status)
}

func TestTimeFormat(*testing.T) {

	var t = time.Now()
	var start = t.Format(time.RFC3339)
	var waitUntil = t.Add(30 * time.Minute).Format(time.RFC3339)
	fmt.Println(start)
	fmt.Println(waitUntil)
}
