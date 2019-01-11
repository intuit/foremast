package metrics

import (
	d "foremast.ai/foremast/foremast-barrelman/pkg/apis/deployment/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"strings"
	"time"
)

//type Interface interface {
//	createMetricsInfo(namespace string, appName string, podNames [][]string, metrics d.Metrics) (MetricsInfo, error)
//}

const CategoryCurrent = "current"
const CategoryBaseline = "baseline"
const CategoryHistorical = "historical"
const StrategyRollingUpdate = "rollingUpdate"
const StrategyCanary = "canary"
const StrategyContinuous = "continuous"

func createMap(namespace string, appName string, podNames []string, metrics d.Metrics, category string, timeWindow time.Duration, strategy string) (map[string]MetricQuery, error) {

	var m = make(map[string]MetricQuery)
	for _, monitoring := range metrics.Monitoring {

		var p = make(map[string]interface{})

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

		var now = time.Now()

		var step int64 = 60
		var nowUnix = (now.Unix() / step) * step
		var before = (now.Add(-timeWindow*time.Minute).Unix() / step) * step

		p["endpoint"] = metrics.Endpoint
		p["step"] = step

		var podCount = len(podNames)

		if category == CategoryCurrent {
			p["start"] = nowUnix
			p["end"] = (now.Add(timeWindow*time.Minute).Unix() / step) * step
			if strategy == StrategyContinuous {
				p["query"] = "namespace_app_per_pod:" + monitoring.MetricName + "{namespace=\"" + namespace + "\",app=\"" + appName + "\"}"
			} else {
				if podCount > 1 {
					p["query"] = "namespace_pod:" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=~\"" + strings.Join(podNames, "|") + "\"}"
				} else {
					p["query"] = "namespace_pod:" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=\"" + podNames[0] + "\"}"
				}
			}
		} else if category == CategoryBaseline {
			p["start"] = before
			p["end"] = nowUnix
			if podCount > 1 {
				p["query"] = "namespace_pod:" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=~\"" + strings.Join(podNames, "|") + "\"}"
			} else {
				p["query"] = "namespace_pod:" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=\"" + podNames[0] + "\"}"
			}
		} else if category == CategoryHistorical {
			//p["step"] = 1200
			var t = now.Add(-7 * 24 * time.Hour)
			p["start"] = t.Unix() / step * step
			p["end"] = nowUnix
			p["query"] = "namespace_app_per_pod:" + monitoring.MetricName + "{namespace=\"" + namespace + "\",app=\"" + appName + "\"}"
		}

		var metricQuery = MetricQuery{
			DataSourceType: metrics.DataSourceType,
			Parameters:     p,
		}

		m[monitoring.MetricAlias] = metricQuery
	}
	return m, nil
}

func CreateMetricsInfo(namespace string, appName string, podNames [][]string, metrics d.Metrics, timeWindow time.Duration, strategy string) (MetricsInfo, error) {
	if len(podNames) == 0 {
		return MetricsInfo{}, errors.NewBadRequest("No valid pod nbames")
	}
	var dataSourceType = metrics.DataSourceType
	if dataSourceType == "prometheus" {

		var podName = []string{}
		if strategy != StrategyContinuous {
			podName = podNames[0]
		}
		var current, err = createMap(namespace, appName, podName, metrics, CategoryCurrent, timeWindow, strategy)
		if err != nil {
			return MetricsInfo{}, nil
		}

		var metricsInfo = MetricsInfo{
			Current: current,
		}

		if strategy != StrategyRollingUpdate && len(podNames) > 1 {
			baseline, err := createMap(namespace, appName, podNames[1], metrics, CategoryBaseline, timeWindow, strategy)
			if err == nil {
				metricsInfo.Baseline = baseline
			}
		}

		historical, err := createMap(namespace, appName, podName, metrics, CategoryHistorical, timeWindow, strategy)
		if err == nil {
			metricsInfo.Historical = historical
		}

		return metricsInfo, nil
	} else {
		return MetricsInfo{}, errors.NewBadRequest("Unsupported DataSourceType:" + dataSourceType)
	}
}

type MetricsInfo struct {
	Current map[string]MetricQuery `json:"current"`

	Baseline map[string]MetricQuery `json:"baseline,omitempty"`

	Historical map[string]MetricQuery `json:"historical,omitempty"`
}

type MetricQuery struct {
	DataSourceType string `json:"dataSourceType"`

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

	Parameters map[string]interface{} `json:"parameters,omitempty"`
}
