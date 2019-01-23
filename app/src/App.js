import React from 'react';
import { withRouter } from 'react-router-dom';
import SplitterLayout from 'react-splitter-layout';
import moment from 'moment';

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
  'namespace_app_per_pod:cpu_usage_seconds_total': [{
    type: BASE,
    name: 'namespace_app_per_pod:cpu_usage_seconds_total',
    tags: '{namespace="default",app="foo"}',
  // },{
  //   type: UPPER,
  //   name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_upper',
  //   tags: '{exported_namespace="foremast-examples",app="foo"}',
  // },{
  //   type: LOWER,
  //   name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_lower',
  //   tags: '{exported_namespace="foremast-examples",app="foo"}',
  // },{
  //   type: ANOMALY,
  //   name: '',
  //   tags: '',
  }],
  'namespace_app_per_pod:memory_usage_bytes': [{
    type: BASE,
    name: 'namespace_app_per_pod:memory_usage_bytes',
    tags: '{namespace="default",app="foo"}',
  // },{
  //   type: UPPER,
  //   name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_upper',
  //   tags: '{exported_namespace="foremast-examples",app="foo"}',
  // },{
  //   type: LOWER,
  //   name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_lower',
  //   tags: '{exported_namespace="foremast-examples",app="foo"}',
  // },{
  //   type: ANOMALY,
  //   name: '',
  //   tags: '',
  }],
  'namespace_app_per_pod:http_server_requests_error_5xx': [{
    type: BASE,
    name: 'namespace_app_per_pod:http_server_requests_error_5xx',
    tags: '{namespace="default",app="foo"}',
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
    name: 'foremastbrain:namespace_pod:http_server_requests_error_5xx_anomaly',
    tags: '{exported_namespace="examples",app="foo"}',
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
};
const dataQueryParam = '?query=';
const dataStartParam = '&start=';
const dataEndParam = '&end=';
const dataStepParam = '&step=';
const dataStepValSec = 60; //data granularity

//API can't provide more than roughly 7 days of data at 60sec granularity
const endTimestamp = moment().subtract(7, 'days').unix();
const startTimestamp = moment().subtract(14, 'days').unix();

class App extends React.Component {
  state = {
    metricName: '',
    baseSeries: {data:[]},
    upperSeries: {data:[]},
    lowerSeries: {data:[]},
    anomalySeries: {data:[]},
  };

  componentDidMount() {
    const pathParam = this.props.match.params.metricName;
    this.setState({metricName: pathParam});

    const queryParams = new URLSearchParams(this.props.location.search);
    const namespaceParam = queryParams.get('namespace') || 'foremast-examples';
    const appNameParam = queryParams.get('app') || 'foo';
    const tagsStr = `{namespace="${namespaceParam}",app="${appNameParam}"}`;
    //TODO:DM - would like to use namespace/app from query params, however, diff series currently use diff tag names (ex: 'namespace' vs 'exported_namespace')

    metricNameMap[pathParam].forEach(metric => {
      let uri = dataDomain + dataPath + dataQueryParam +
        encodeURIComponent(metric.name + metric.tags) +
        dataStartParam + startTimestamp + dataEndParam + endTimestamp +
        dataStepParam + dataStepValSec;
      let responsePromise = fetch(uri);
      switch (metric.type) {
        case BASE:
          responsePromise.then(resp => this.processBaseResponse(resp));
          break;
        case UPPER:
          responsePromise.then(resp => this.processUpperResponse(resp));
          break;
        case LOWER:
          responsePromise.then(resp => this.processLowerResponse(resp));
          break;
        case ANOMALY:
          responsePromise.then(resp => {
            //TODO:DM - this is a hack to ensure that the base series is loaded before attempting to process anomalies; instead should use promose resolution to signal ready to process anomalies
            setTimeout(this.processAnomalyResponse.bind(this, resp), 1000);
          });
          break;
        default:
          break;
      }
    });
  }

  render() {
    let { metricName, baseSeries, upperSeries,
      lowerSeries, anomalySeries } = this.state;
    return (
      <div className="App">
        <SplitterLayout vertical={true}>
          <TimeseriesChart
            metricName={metricName}
            baseSeries={baseSeries}
            upperSeries={upperSeries}
            lowerSeries={lowerSeries}
            anomalySeries={anomalySeries}
          />
          <ScatterChart/>
        </SplitterLayout>
      </div>
    );
  }

  //TODO:DM - how to clean-up copy/paste of next 3 fns?
  processBaseResponse(resp) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      this.setState({baseSeries: {name, data}});
    });

  }
  processUpperResponse(resp) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      this.setState({upperSeries: {name, data}});
    });
  }
  processLowerResponse(resp) {
    this.processResponse(resp).then(result => {
      let data = result.values.map(point => [1000 * point[0], parseFloat(point[1])]);
      let name = (result.metric ? result.metric.__name__ : null);
      this.setState({lowerSeries: {name, data}});
    });
  }
  processAnomalyResponse(resp) {
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
        this.state.baseSeries.data.forEach(basePoint => {
          let timeDiff = anomalyTimestamp - basePoint[0];
          //use this point if it's within a minute (data resolution requested), but only if BEFORE anomaly stamp
          if(timeDiff < dataStepValSec * 1000 && timeDiff > 0) {
            //NOTE: using base point here will allow for anomolous points to fall directly on top of measured series BUT does therefore indicate slightly different timing than the anaomalies may be marked with
            //NOTE: also, this strategy allows for out of order points to be added, highcharts will warn about this with error #15, but it doesn't stop it from rendering as expected
            data.push(basePoint);
          }
        });
      });
      this.setState({anomalySeries: {
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
      }});
    });
  }
  processResponse(resp) {
    return new Promise((resolve, reject) => {
      if (resp.ok) {
        resp.json().then(respStr => {
          let tempParsed = JSON.parse(respStr);
          let tempResult = tempParsed.data.result.length ? tempParsed.data.result[0] : {values: []};
          resolve(tempResult);
        });
      } else {
        reject('Response object not OK');
      }
    });
  }
}

export default withRouter(App);