import React from 'react';
import { withRouter } from 'react-router-dom';
import SplitterLayout from 'react-splitter-layout';
import moment from 'moment';
import Highcharts from 'highcharts';

import './App.css';
import TimeseriesChart from './components/charts/timeseries/TimeseriesChart';
import ScatterChart from './components/charts/scatter/ScatterChart';

const BASE = 'base';
const UPPER = 'upper';
const LOWER = 'lower';
const ANOMALY = 'anomaly';

const dataDomain = 'http://foremast-api-service.foremast.svc.cluster.local:8099';
const dataPath = '/api/v1/query_range';

const metricNameMap = {
  'namespace_app_per_pod:http_server_requests_error_5xx': [{
    type: BASE,
    name: 'namespace_app_per_pod:http_server_requests_error_5xx',
    tags: '{namespace="foremast-examples",app="foo"}',
  },{
    type: UPPER,
    name: 'foremastbrain:namespace_app_per_pod:http_server_requests_error_5xx_upper',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: LOWER,
    name: 'foremastbrain:namespace_app_per_pod:http_server_requests_error_5xx_lower',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: ANOMALY,
    name: 'foremastbrain:namespace_app_per_pod:http_server_requests_error_5xx_anomaly',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  }],
  'namespace_app_per_pod:http_server_requests_latency': [{
    type: BASE,
    name: 'namespace_app_per_pod:http_server_requests_latency',
    tags: '{namespace="foremast-examples",app="foo"}',
  },{
    type: UPPER,
    name: 'foremastbrain:namespace_app_per_pod:http_server_requests_latency_upper',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: LOWER,
    name: 'foremastbrain:namespace_app_per_pod:http_server_requests_latency_lower',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: ANOMALY,
    name: 'foremastbrain:namespace_app_per_pod:http_server_requests_latency_anomaly',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  }],
  'namespace_app_per_pod:cpu_usage_seconds_total': [{
    type: BASE,
    name: 'namespace_app_per_pod:cpu_usage_seconds_total',
    tags: '{namespace="foremast-examples",app="foo"}',
  },{
    type: UPPER,
    name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_upper',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: LOWER,
    name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_lower',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: ANOMALY,
    name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_anomaly',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  }],
  'namespace_app_per_pod:memory_usage_bytes': [{
    type: BASE,
    name: 'namespace_app_per_pod:memory_usage_bytes',
    tags: '{namespace="foremast-examples",app="foo"}',
  },{
    type: UPPER,
    name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_upper',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: LOWER,
    name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_lower',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  },{
    type: ANOMALY,
    name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_anomaly',
    tags: '{exported_namespace="foremast-examples",app="foo"}',
  }],
  'sum by (label_version) (kube_pod_labels{label_app="foo", namespace="foremast-examples"})': [{
    type: BASE,
    name: 'sum by (label_version) (kube_pod_labels{label_app="foo", namespace="foremast-examples"})',
    tags: ''
  }]
};
const dataQueryParam = '?query=';
const dataStartParam = '&start=';
const dataEndParam = '&end=';
const dataStepParam = '&step=';
const dataStepValSec = 15; //data granularity

let seriesTemplate = {
  'namespace_app_per_pod:http_server_requests_latency':{data:[]},
  'namespace_app_per_pod:http_server_requests_error_5xx':{data:[]},
  'namespace_app_per_pod:cpu_usage_seconds_total':{data:[]},
  'namespace_app_per_pod:memory_usage_bytes':{data:[]},
  'sum by (label_version) (kube_pod_labels{label_app="foo", namespace="foremast-examples"})':{data:[]}
};

class App extends React.Component {
  state = {
    baseSeries: {...seriesTemplate},
    upperSeries: {...seriesTemplate},
    lowerSeries: {...seriesTemplate},
    anomalySeries: {...seriesTemplate},
    xSeries: {data:[]},
    ySeries: {data:[]},
  };

  componentDidMount() {
    ["mousemove", "touchmove", "touchstart"].forEach(function (eventType) {
      document
        .getElementById("container")
        .addEventListener(eventType, function (e) {
          var chart, point, i, event;

          for (i = 0; i < Highcharts.charts.length; i = i + 1) {
            chart = Highcharts.charts[i];
            //filter out scatter charts... or handle them differently
            if (chart) {// && chart.options.chart.type === 'line') {
              // Find coordinates within the chart
              event = chart.pointer.normalize(e);
              // Get the hovered point
              point = chart.series[0].searchPoint(event, true);
              if (point) {
                point.highlight(e);
              }
            }
          }
        });
    });
    this.fetchData();
    setInterval(this.fetchData, dataStepValSec * 1000);
  }

  fetchData = () => {
    //API can't provide more than roughly 7 days of data at 60sec granularity
    const endTimestamp = moment().subtract(0, 'minutes').unix();
    const startTimestamp = moment().subtract(20, 'minutes').unix();

    // const pathParam = this.props.match.params.metricName;
    // this.setState({metricName: pathParam});
    //
    // const queryParams = new URLSearchParams(this.props.location.search);
    // const namespaceParam = queryParams.get('namespace') || 'foremast-examples';
    // const appNameParam = queryParams.get('app') || 'foo';
    // const tagsStr = `{namespace="${namespaceParam}",app="${appNameParam}"}`;
    //TODO:DM - would like to use namespace/app from query params, however, diff series currently use diff tag names (ex: 'namespace' vs 'exported_namespace')

    Object.keys(metricNameMap).forEach(key => {
      metricNameMap[key].forEach(metric => {
        let uri = dataDomain + dataPath + dataQueryParam +
          encodeURIComponent(metric.name + metric.tags) +
          dataStartParam + startTimestamp + dataEndParam + endTimestamp +
          dataStepParam + dataStepValSec;
        let responsePromise = fetch(uri);
        switch (metric.type) {
          case BASE:
            responsePromise.then(resp => this.processBaseResponse(resp, key));
            break;
          case UPPER:
            responsePromise.then(resp => this.processUpperResponse(resp, key));
            break;
          case LOWER:
            responsePromise.then(resp => this.processLowerResponse(resp, key));
            break;
          case ANOMALY:
            responsePromise.then(resp => {
              //TODO:DM - this is a hack to ensure that the base series is loaded before attempting to process anomalies; instead should use promise resolution to signal ready to process anomalies
              setTimeout(this.processAnomalyResponse.bind(this, resp, key), 1000);
            });
            break;
          default:
            break;
        }
      });
    });

    //now force requests for Mem v CPU
    let cpuMetric = metricNameMap['namespace_app_per_pod:cpu_usage_seconds_total'][0];
    let cpuUri = dataDomain + dataPath + dataQueryParam +
      encodeURIComponent(cpuMetric.name + cpuMetric.tags) +
      dataStartParam + startTimestamp + dataEndParam + endTimestamp +
      dataStepParam + dataStepValSec;
    fetch(cpuUri)
      .then(resp => this.processYResponse(resp));

    let memMetric = metricNameMap['namespace_app_per_pod:memory_usage_bytes'][0];
    let memUri = dataDomain + dataPath + dataQueryParam +
      encodeURIComponent(memMetric.name + memMetric.tags) +
      dataStartParam + startTimestamp + dataEndParam + endTimestamp +
      dataStepParam + dataStepValSec;
    fetch(memUri)
      .then(resp => this.processXResponse(resp));
  };

  render() {
    let { baseSeries, upperSeries, lowerSeries,
      anomalySeries, xSeries, ySeries } = this.state;
    return (
      <div className="App">
        <SplitterLayout vertical={true}>
          <div id="container">
            {
              Object.keys(metricNameMap).map(key => {
                return (
                  <TimeseriesChart
                    key={key}
                    metricName={key}
                    baseSeries={baseSeries[key]}
                    upperSeries={upperSeries[key]}
                    lowerSeries={lowerSeries[key]}
                    anomalySeries={anomalySeries[key]}
                  />
                );
              })
            }
          </div>
          <ScatterChart
            xSeries={xSeries}
            ySeries={ySeries}
          />
        </SplitterLayout>
      </div>
    );
  }

  //TODO:DM - how to clean-up copy/paste of next 5 fns? notice sometimes diff scaling
  processXResponse(resp) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])/1000000]);
      let name = (result.metric ? result.metric.__name__ : null);
      this.setState({xSeries: {name, data}});
    });

  }
  processYResponse(resp) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], 100 * parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      this.setState({ySeries: {name, data}});
    });

  }
  processBaseResponse(resp, key) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      let newState = {
          ...this.state.baseSeries,
          [key]: {name, data}
        };
      this.setState({baseSeries: newState});
    });

  }
  processUpperResponse(resp, key) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      let newState = {
        ...this.state.upperSeries,
        [key]: {name, data}
      };
      this.setState({upperSeries: newState});
    });
  }
  processLowerResponse(resp, key) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      let newState = {
        ...this.state.lowerSeries,
        [key]: {name, data}
      };
      this.setState({lowerSeries: newState});
    });
  }
  processAnomalyResponse(resp, key) {
    this.processResponse(resp).then(result => {
      let data = [];
      let name = (result.metric ? result.metric.__name__ : null);
      let anomalyArr = result.values.map(point => 1000 * parseInt(point[1]));
      let seen = new Set();
      anomalyArr.forEach(anomalyTimestamp => {
        seen.add(anomalyTimestamp);
      });
      let uniqueAnomalyTimestamps = [];
      for (let time of seen.keys()) {
        //NOTE: this presumes chronological ordering in original data response; not sure if that is warranted or should be explicitly sorted here
        uniqueAnomalyTimestamps.push(time);
      }
      //TODO: DM any way to clean up this n * m processing; for each anomaly timestamp, see if it exists in the base series and use its y value, if so
      uniqueAnomalyTimestamps.forEach(anomalyTimestamp => {
        this.state.baseSeries[key].data.forEach(basePoint => {
          let timeDiff = anomalyTimestamp - basePoint[0];
          //use this point if it's within a minute (data resolution requested), but only if BEFORE anomaly stamp
          if(timeDiff < dataStepValSec * 1000 && timeDiff > 0) {
            //NOTE: using base point here will allow for anomolous points to fall directly on top of measured series BUT does therefore indicate slightly different timing than the anaomalies may be marked with
            //NOTE: also, this strategy allows for out of order points to be added, highcharts will warn about this with error #15, but it doesn't stop it from rendering as expected
            data.push(basePoint);
          }
        });
      });
      let newState = {
        ...this.state.anomalySeries,
        [key]: {
          name,
          data,
          color: '#FF0000',
          marker: {
            enabled: true,
            symbol: 'circle'
          },
          lineWidth: 0,
          states: {
            hover: {
              lineWidthPlus: 0
            }
          }
        }
      };
      this.setState({anomalySeries: newState});
    });
  }
  processResponse(resp) {
    return new Promise((resolve, reject) => {
      if (resp.ok) {
        resp.json().then(respStr => {
          let tempParsed = JSON.parse(respStr);
          //TODO:DM - rather than grabbing last item, iterate thru and build combined arr?
          let tempResult = tempParsed.data.result.length ?
            tempParsed.data.result[tempParsed.data.result.length - 1] : {values: []};
          resolve(tempResult);
        });
      } else {
        reject('Response object not OK');
      }
    });
  }
}

export default withRouter(App);