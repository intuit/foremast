import React from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import PropTypes from 'prop-types';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';

import * as highlightActions from '../../../actions/highlightActions';

class ScatterChart extends React.Component {
  render() {
    let options = {
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
          //TODO:DM - make x and y axis titles change dynamically based on metric unit
          text: 'Memory (MB)'
        }
      },
      yAxis: {
        title: {
          text: 'CPU %'
        }
      },
      series: this.buildSeries(),
      legend: {
        enabled: false
      },
      credits: {
        enabled: false
      },
    };

    return (
      <HighchartsReact
        highcharts={Highcharts}
        options={options}
      />
    );
  }
  buildSeries() {
    let xytData = [];
    const { xSeries, ySeries, highlight } = this.props;
    let timeDiffMs = +Infinity;
    //series lengths may differ as they are loaded asynchronously, only build series once they are the same length
    if (xSeries.data.length === ySeries.data.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < xSeries.data.length; i++){
        timeDiffMs = highlight.timestamp - xSeries.data[i][0];
        // xytData.push({
        //   x: xSeries.data[i][1],
        //   y: ySeries.data[i][1],
        //   selected: timeDiffMs > 0 && timeDiffMs < 60 * 1000,
        //   timestamp: xSeries.data[i][0]
        // });
        xytData.push([
          xSeries.data[i][1],
          ySeries.data[i][1],
          xSeries.data[i][0]
        ]);
      }
    }
    let xytSeries = {
      type: 'scatter',
      name: 'CPU vs Memory',
      data: xytData,
      marker: {
        radius: 2
      },
      tooltip: {
        followPointer: false,
        pointFormat: '[{point.x:.5f}, {point.y:.5f}]'
      }
    };
    return [xytSeries];
  }
}

ScatterChart.propTypes = {
  highlightActions: PropTypes.object,
  highlight: PropTypes.object
};

const mapStoreToProps = store => ({highlight: store.highlight});

const mapDispatchToProps = dispatch => ({
  highlightActions: bindActionCreators(highlightActions, dispatch)
});

export default connect(
  mapStoreToProps,
  mapDispatchToProps
)(ScatterChart);