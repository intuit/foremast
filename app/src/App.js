import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';
import Highcharts from 'highcharts';
import HighchartsMore from 'highcharts/highcharts-more'; //necessary for 'arearange' series type
import Highstock from 'highcharts/highstock';
import HighchartsReact from 'highcharts-react-official';
import moment from 'moment';

import './App.css';

HighchartsMore(Highstock);

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
    tags: '{namespace="default",app="foo"}',
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
const dataStepValSec = 60; //60 second granularity

//API can't provide more than roughly 7 days of data at 60sec granularity
const endTimestamp = moment().subtract(5, 'days').unix();
const startTimestamp = moment().subtract(12, 'days').unix();

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      metricName: '',
      baseSeries: {data:[]},
      upperSeries: {data:[]},
      lowerSeries: {data:[]},
      anomalySeries: {data:[]},
    };
  }

  componentDidMount() {
    //TODO:DM - read path params to provide metric name
    const pathComponents = this.props.location.pathname.split('/');
    //TODO:DM - simply grabbing last param after '/' feels fragile
    //could be empty string... a better default to use, if so?
    const pathParam = pathComponents[pathComponents.length - 1];
    this.setState({metricName: pathParam});

    metricNameMap[pathParam].forEach(metric => {
      let uri = dataDomain + dataPath + dataQueryParam +
        encodeURIComponent(metric.name + metric.tags) + dataStartParam +
        startTimestamp + dataEndParam + endTimestamp + dataStepParam +
        dataStepValSec;
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
          responsePromise.then(resp => this.processAnomalyResponse(resp));
          break;
        default:
          break;
      }
    });
  }

  render() {
    let dynamicTimeseriesOptions = {
      ...timeseriesOptions,
      series: this.buildHighchartsSeries(),
      title: {
        text: this.state.metricName + ' Metric + Modeled Range',
      },
    };
    return (
      <div className="App">
        <HighchartsReact
          highcharts={Highstock}
          constructorType={'stockChart'}
          options={dynamicTimeseriesOptions}
        />
        <HighchartsReact
          highcharts={Highcharts}
          options={scatterOptions}
        />
      </div>
    );
  }

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
          if(anomalyTimestamp === basePoint[0]) {
            data.push(basePoint);
          }
        });
      });
      this.setState({anomalySeries: {
        name,
        data,
        marker: {
          symbol: 'url(img/explosion.png)'
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
  buildHighchartsSeries() {
    let rangeData = [];
    if (this.state.upperSeries.data.length === this.state.lowerSeries.data.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < this.state.upperSeries.data.length; i++){
        rangeData.push([
          this.state.upperSeries.data[i][0],
          this.state.lowerSeries.data[i][1],
          this.state.upperSeries.data[i][1]
        ]);
      }
    }
    let rangeSeries = {
      type: 'arearange',
      showInLegend: false,
      name: 'Model Range',
      data: rangeData,
      fillOpacity: 0.1
    };
    return [this.state.baseSeries, this.state.anomalySeries, rangeSeries];
  }
}

export default withRouter(App);

const timeseriesOptions = {
  chart: {
    zoomType: 'x'
  },

  subtitle: {
    text: document.ontouchstart === undefined ?
      'Click and drag in the plot area to zoom in' : 'Pinch the chart to zoom in'
  },
  xAxis: {
    type: 'datetime'
  },
  yAxis: {
    title: {
      //how to change this depending on data loaded?
      text: 'Seconds'
    }
  },
  legend: {
    layout: 'vertical',
    align: 'right',
    verticalAlign: 'middle'
  },

  plotOptions: {
    series: {
      label: {
        connectorAllowed: false
      },
    }
  },

  responsive: {
    rules: [{
      condition: {
        maxWidth: 500
      },
      chartOptions: {
        legend: {
          layout: 'horizontal',
          align: 'center',
          verticalAlign: 'bottom'
        }
      }
    }]
  },

  rangeSelector: {
    buttons: [{
      type: 'day',
      count: 1,
      text: '1d',
    }, {
      type: 'day',
      count: 3,
      text: '3d',
    }, {
      type: 'week',
      count: 1,
      text: '1w',
    // }, {
    //   type: 'week',
    //   count: 2,
    //   text: '2w',
    // }, {
    //   type: 'month',
    //   count: 1,
    //   text: '1m',
    // }, {
    //   type: 'month',
    //   count: 3,
    //   text: '3m',
    // }, {
    //   type: 'month',
    //   count: 6,
    //   text: '6m',
    // }, {
    //   type: 'ytd',
    //   text: 'ytd',
    // }, {
    //   type: 'year',
    //   count: 1,
    //   text: '1y',
    }, {
      type: 'all',
      text: 'all',
    }],
  },
  credits: {
    enabled: false
  },
};

const scatterOptions = {
  chart: {
    type: 'scatter',
    zoomType: 'xy'
  },

  title: {
    text: 'CPU vs Memory'
  },

  subtitle: {
    text: document.ontouchstart === undefined ?
      'Click and drag in the plot area to zoom in' : 'Pinch the chart to zoom in'
  },
  xAxis: {
    title: {
      text: 'Memory (MB)'
    }
  },
  yAxis: {
    title: {
      text: 'CPU %'
    }
  },
  legend: {
    enabled: false
  },

  series: [{
    name: 'CpuVsMemory',
    data: [[512000,0.7537],[568000,0.7547],[594000,0.7559],[506000,0.7631],[518000,0.7644],[400000,0.569],[812000,0.9683],[502000,0.77],[532000,0.7703],[555000,0.7057],[500000,0.69728],[504000,0.7721],[517000,0.7748],[519000,0.774],[523000,0.7718],[492000,0.7731],[510000,0.767],[511000,0.769],[512500,0.7706],[506000,0.7752],[512000,0.874]]
  }],
  credits: {
    enabled: false
  },
};