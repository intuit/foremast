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
const StrategyHpa = "hpa"

func indexOf(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

func createMap(namespace string, appName string, podNames []string, metrics d.Metrics, category string,
	timeWindow time.Duration, strategy string, metricsAliases []string) (map[string]MetricQuery, error) {

	var m = make(map[string]MetricQuery)
	for i, monitoring := range metrics.Monitoring {

		var priority = i + 1
		if metricsAliases != nil {
			var index = indexOf(metricsAliases, monitoring.MetricAlias)
			if index == -1 {
				continue
			}
			priority = index + 1
		}

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
			//Since prometheus has 1 minutes latency, add one minute to make sure prometheus won't get metrics from previous version
			p["start"] = nowUnix + step
			p["end"] = (now.Add((timeWindow+1)*time.Minute).Unix() / step) * step
			if strategy == StrategyContinuous || strategy == StrategyHpa {
				p["query"] = "namespace_app_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",app=\"" + appName + "\"}"
			} else {
				if podCount > 1 {
					p["query"] = "namespace_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=~\"" + strings.Join(podNames, "|") + "\"}"
				} else {
					p["query"] = "namespace_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=\"" + podNames[0] + "\"}"
				}
			}
		} else if category == CategoryBaseline {
			p["start"] = before
			p["end"] = nowUnix
			if podCount > 1 {
				p["query"] = "namespace_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=~\"" + strings.Join(podNames, "|") + "\"}"
			} else {
				p["query"] = "namespace_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",pod=\"" + podNames[0] + "\"}"
			}
		} else if category == CategoryHistorical {
			//p["step"] = 1200
			var t = now.Add(-7 * 24 * time.Hour)
			p["start"] = t.Unix() / step * step
			p["end"] = nowUnix
			p["query"] = "namespace_app_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",app=\"" + appName + "\"}"
		}

		var metricQuery = MetricQuery{
			DataSourceType: metrics.DataSourceType,
			Priority:       priority,
			Parameters:     p,
		}

		m[monitoring.MetricAlias] = metricQuery
	}
	return m, nil
}

func CreateMetricsInfo(namespace string, appName string, podNames [][]string, metrics d.Metrics, timeWindow time.Duration, strategy string, metricAliases []string) (MetricsInfo, error) {
	if !(strategy == StrategyContinuous || strategy == StrategyHpa) && len(podNames) == 0 {
		return MetricsInfo{}, errors.NewBadRequest("No valid pod names")
	}
	var dataSourceType = metrics.DataSourceType
	if dataSourceType == "prometheus" {
		var podName = []string{}
		if !(strategy == StrategyContinuous || strategy == StrategyHpa) {
			podName = podNames[0]
		}
		var current, err = createMap(namespace, appName, podName, metrics, CategoryCurrent, timeWindow, strategy, metricAliases)
		if err != nil {
			return MetricsInfo{}, nil
		}

		var metricsInfo = MetricsInfo{
			Current: current,
		}

		if strategy != StrategyRollingUpdate && len(podNames) > 1 {
			baseline, err := createMap(namespace, appName, podNames[1], metrics, CategoryBaseline, timeWindow, strategy, metricAliases)
			if err == nil {
				metricsInfo.Baseline = baseline
			}
		}

		historical, err := createMap(namespace, appName, podName, metrics, CategoryHistorical, timeWindow, strategy, metricAliases)
		if err == nil {
			metricsInfo.Historical = historical
		}

		return metricsInfo, nil
	} else {
		return MetricsInfo{}, errors.NewBadRequest("Unsupported DataSourceType:" + dataSourceType)
	}
}

func CreatePodCountURL(namespace string, appName string, metrics d.Metrics, timeWindow time.Duration) (MetricQuery, error) {
	for _, monitoring := range metrics.Monitoring {
		if monitoring.MetricAlias == "count" {
			var now = time.Now()
			var p = make(map[string]interface{})
			var step int64 = 60
			var nowUnix = (now.Unix() / step) * step
			p["endpoint"] = metrics.Endpoint
			p["step"] = step
			p["start"] = nowUnix + step
			p["end"] = (now.Add((timeWindow+1)*time.Minute).Unix() / step) * step
			p["query"] = "namespace_app_pod_" + monitoring.MetricName + "{namespace=\"" + namespace + "\",app=\"" + appName + "\"}"
			var metricQuery = MetricQuery{
				DataSourceType: metrics.DataSourceType,
				Parameters:     p,
			}
			return metricQuery, nil
		}
	}
	return MetricQuery{}, errors.NewBadRequest("No count metric found:" + appName)
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
	Priority int `json:"priority,omitempty"`

	Parameters map[string]interface{} `json:"parameters,omitempty"`
}
