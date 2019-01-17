import React, { Component } from 'react';
import { withRouter } from 'react-router-dom';
import Highcharts from 'highcharts';
import Highstock from 'highcharts/highstock';
import HighchartsReact from 'highcharts-react-official';

import './App.css';

var moment = require('moment');

const dataDomain = 'http://foremast-api-service.foremast.svc.cluster.local:8099';
const dataPath = '/api/v1/query_range';
const metricNameMap = {
  'namespace_app_per_pod:cpu_usage_seconds_total': [
    'namespace_app_per_pod:cpu_usage_seconds_total'
  ],
  'namespace_app_per_pod:memory_usage_bytes': [
    'namespace_app_per_pod:memory_usage_bytes'
  ],
  'namespace_app_per_pod:http_server_requests_error_5xx': [
    'namespace_app_per_pod:http_server_requests_error_5xx',
    //'namespace_app_per_pod:http_server_requests_error_5xx_upper',
    //'namespace_app_per_pod:http_server_requests_error_5xx_lower',
    //'namespace_app_per_pod:http_server_requests_error_5xx_anomaly'
  ],
  'namespace_app_per_pod:http_server_requests_latency': [
    'namespace_app_per_pod:http_server_requests_latency'
  ],
};
const dataQueryParamsA = '?query=';
const dataQueryParamsB = '{namespace="default",app="foo"}';
const dataQueryParamsC = '&start=';
const dataQueryParamsD = '&end=';
const dataQueryParamsE = '&step=60';

//API can't provide more than roughly 7 days of data at 60sec granularity
const endTimestamp = moment().subtract(0, 'days').unix();
const startTimestamp = moment().subtract(7, 'days').unix();

function makeResponse(resp) {
  if (resp.ok) {
    return resp.json();
  } else {
    //what else to do here?
    return null;
  }
}

class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      series: []
    };
  }

  componentDidMount() {
    //TODO:DM - read path params to provide metric name
    const pathComponents = this.props.location.pathname.split('/');
    //TODO:DM - simply grabbing last param after '/' feels fragile
    //could be empty string... a better default to use, if so?
    const pathParam = pathComponents[pathComponents.length - 1];

    metricNameMap[pathParam].forEach(metricName => {
      let uri = dataDomain + dataPath + dataQueryParamsA +
        encodeURIComponent(metricName + dataQueryParamsB) + dataQueryParamsC + startTimestamp + dataQueryParamsD +
        endTimestamp + dataQueryParamsE;
      fetch(uri)
      .then(resp => makeResponse(resp))
      //TODO:DM - make following anon fn into named and possibly combine with makeResponse
      .then(data => {
        let tempParsed = JSON.parse(data);
        let tempResult = tempParsed.data.result.length ? tempParsed.data.result[0] : {values: []};
        let tempData = tempResult.values.map(point => [1000 * point[0], parseFloat(point[1])]);
        let tempName = tempResult.metric.__name__;
        this.setState({
          series: [{name: tempName, data: tempData}]
        });
      });
    });
  }

  render() {
    let dynamicTimeseriesOptions = {
      ...timeseriesOptions,
      series: this.state.series
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
}

export default withRouter(App);

const timeseriesOptions = {
  chart: {
    zoomType: 'x'
  },

  title: {
    text: 'Error5xxCount Metric + Model'
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
      text: 'Number of Errors'
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

  // series: [{
  //   name: 'Error5xxCountUpper',
  //   data: [[1539120304467,0.8537],[1539120374467,0.8537],[1539120374467,0.869],[1539120444467,0.869],[1539120444467,0.8728],[1539120514467,0.8728],[1539120514467,0.8731],[1539120584467,0.8731],[1539120584467,0.874]]
  // }, {
  //   name: 'Error5xxCount',
  //   data: [[1539120304467,0.7537],[1539120318467,0.7537],[1539120332467,0.7559],[1539120346467,0.7631],[1539120360467,0.9644],[1539120374467,0.769],[1539120388467,0.7683],[1539120402467,0.77],[1539120416467,0.7703],[1539120430467,0.7757],[1539120444467,0.7728],[1539120458467,0.7721],[1539120472467,0.7748],[1539120486467,0.574],[1539120500467,0.7718],[1539120514467,0.7731],[1539120528467,0.767],[1539120542467,0.769],[1539120556467,0.7706],[1539120570467,0.7752],[1539120584467,0.774]]
  // }, {
  //   name: 'Error5xxCountLower',
  //   data: [[1539120304467,0.6537],[1539120374467,0.669],[1539120444467,0.6728],[1539120514467,0.6731],[1539120584467,0.674]]
  // }, {
  //   name: 'Error5xxCountAnomaly',
  //   data: [[1539120360467,0.9644],[1539120486467,0.574]],
  //   marker: {
  //     symbol: 'url(img/explosion.png)'
  //   },
  //   lineWidth: 0,
  //   states: {
  //     hover: {
  //       lineWidthPlus: 0
  //     }
  //   }
  // }],

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
    //allButtonsEnabled: true,
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
    }, {
      type: 'week',
      count: 2,
      text: '2w',
    }, {
      type: 'month',
      count: 1,
      text: '1m',
    }, {
      type: 'month',
      count: 3,
      text: '3m',
    }, {
      type: 'month',
      count: 6,
      text: '6m',
    }, {
      type: 'ytd',
      text: 'ytd',
    }, {
      type: 'year',
      count: 1,
      text: '1y',
    }, {
      type: 'all',
      text: 'all',
    }],
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
};