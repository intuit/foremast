import React from 'react';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';

class ScatterChart extends React.Component {
  render() {
    //let { xName, xSeries, yName, ySeries } = this.state;
    let options = {
      ...scatterOptions,
      //series: this.buildHighchartsSeries(),
      // title: {
      //   text: this.props.metricName + ' Metric + Modeled Range',
      // },
    };
    return (
      <HighchartsReact
        highcharts={Highcharts}
        options={options}
      />
    );
  }
}

export default ScatterChart;

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