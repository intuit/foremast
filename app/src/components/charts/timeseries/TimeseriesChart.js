import React from 'react';
import HighchartsMore from 'highcharts/highcharts-more'; //necessary for 'arearange' series type
import Highstock from 'highcharts/highstock';
import HighchartsReact from 'highcharts-react-official';

HighchartsMore(Highstock);

class TimeseriesChart extends React.Component {
  render() {
    let options = {
      ...timeseriesOptions,
      series: this.buildHighchartsSeries(),
      title: {
        text: this.props.metricName + ' Metric + Modeled Range',
      },
    };
    return (
      <HighchartsReact
        highcharts={Highstock}
        constructorType={'stockChart'}
        options={options}
      />
    );
  }

  buildHighchartsSeries() {
    let rangeData = [];
    const { baseSeries, upperSeries, lowerSeries, anomalySeries } = this.props;
    if (upperSeries.data.length === lowerSeries.data.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < upperSeries.data.length; i++){
        rangeData.push([
          upperSeries.data[i][0],
          lowerSeries.data[i][1],
          upperSeries.data[i][1]
        ]);
      }
    } else {
      console.warn('Upper and Lower series of different lengths, cannot build range series.')
    }
    let rangeSeries = {
      type: 'arearange',
      showInLegend: false,
      name: 'Model Range',
      data: rangeData,
      fillOpacity: 0.1
    };
    return [baseSeries, anomalySeries, rangeSeries];
  }
}

export default TimeseriesChart;

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
      //TODO:DM - change this based on metric name selected, may need to hard-code unit with rest of data in App.js
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