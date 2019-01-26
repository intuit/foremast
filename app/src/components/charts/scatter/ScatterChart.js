import React from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import PropTypes from 'prop-types';
import Highcharts from 'highcharts';
import Highcharts3d from 'highcharts/highcharts-3d'; //necessary for 'arearange' series type
import HighchartsReact from 'highcharts-react-official';

import * as highlightActions from '../../../actions/highlightActions';

Highcharts3d(Highcharts);

setTimeout(function () {
  let chart;

  function dragStart(chart) {
    return (eStart) => {
      eStart = chart.pointer.normalize(eStart);

      var posX = eStart.chartX,
          posY = eStart.chartY,
          alpha = chart.options.chart.options3d.alpha,
          beta = chart.options.chart.options3d.beta,
          sensitivity = 5,  // lower is more sensitive
          handlers = [];

      function drag(e) {
          // Get e.chartX and e.chartY
          e = chart.pointer.normalize(e);

          chart.update({
              chart: {
                  options3d: {
                      alpha: alpha + (e.chartY - posY) / sensitivity,
                      beta: beta + (posX - e.chartX) / sensitivity
                  }
              }
          }, undefined, undefined, false);
      }

      function unbindAll() {
          handlers.forEach(function (unbind) {
              if (unbind) {
                  unbind();
              }
          });
          handlers.length = 0;
      }

      handlers.push(Highcharts.addEvent(document, 'mousemove', drag));
      handlers.push(Highcharts.addEvent(document, 'touchmove', drag));
      handlers.push(Highcharts.addEvent(document, 'mouseup', unbindAll));
      handlers.push(Highcharts.addEvent(document, 'touchend', unbindAll));
    };

  }

  for (let i = 0; i < Highcharts.charts.length; i = i + 1) {
    chart = Highcharts.charts[i];
    if(chart.options.chart.type === 'scatter3d'){
      Highcharts.addEvent(chart.container, 'mousedown', dragStart(chart));
      Highcharts.addEvent(chart.container, 'touchstart', dragStart(chart));

    }
  }
}, 1000);

// Give the points a 3D feel by adding a radial gradient
Highcharts.setOptions({
    colors: Highcharts.getOptions().colors.map(function (color) {
        return {
            radialGradient: {
                cx: 0.4,
                cy: 0.3,
                r: 0.5
            },
            stops: [
                [0, color],
                [1, Highcharts.Color(color).brighten(-0.2).get('rgb')]
            ]
        };
    })
});

class ScatterChart extends React.Component {
  render() {
    let options = {
      chart: {
        type: 'scatter3d',
        animation: false,
        options3d: {
          enabled: true,
          //beta: 90,
          depth: 750,
          viewDistance: 5,
          fitToPlot: false,
          frame: {
            bottom: { size: 1, color: 'rgba(0,0,0,0.02)' },
            back: { size: 1, color: 'rgba(0,0,0,0.04)' },
            side: { size: 1, color: 'rgba(0,0,0,0.06)' }
          }
        }
      },
      title: {
        text: 'Time vs CPU vs Memory'
      },
      // subtitle: {
      //   text: document.ontouchstart === undefined ?
      //     'Click and drag in the plot area to zoom in' : 'Pinch the chart to zoom in'
      // },
      xAxis: {
        title: {
          text: 'Time'
        },
        type: 'datetime'
      },
      yAxis: {
        title: {
          text: 'CPU %'
        }
      },
      zAxis:{
        title: {
          text: 'Memory (MB)'
        }
      },
      tooltip: {
        formatter: function() {
          return  ('<b>' + this.series.name +'</b><br/>' +
            Highcharts.dateFormat('%e%b%Y %H:%M:%S', new Date(this.x)) +
            '<br/>CPU: ' + this.y.toFixed(5) + ' %' +
            '<br/>Memory:' + this.point.z.toFixed(2) + ' MB');
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
        id="scatter"
        highcharts={Highcharts}
        options={options}
      />
    );
  }
  buildSeries() {
    let xytData = [];
    let { xSeries, ySeries } = this.props;
    //immediately enforce the same lengths
    if(xSeries.data.length !== 0 && ySeries.data.length !== 0){
      xSeries.data.length = Math.min(xSeries.data.length, ySeries.data.length);
      ySeries.data.length = Math.min(xSeries.data.length, ySeries.data.length);
    }
    //NOTE: this causes us to either LOSE data (from longer series) or create data (for shorter series)
    //let timeDiffMs = +Infinity;
    //series lengths may differ as they are loaded asynchronously, only build series once they are the same length
    if (xSeries.data.length === ySeries.data.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < xSeries.data.length; i++){
        //timeDiffMs = highlight.timestamp - xSeries.data[i][0];
        xytData.push([
          xSeries.data[i][0],
          ySeries.data[i][1],
          xSeries.data[i][1],
        ]);
      }
    }
    let xytSeries = {
      type: 'scatter3d',
      name: 'Memory vs CPU',
      data: xytData,
      marker: {
        radius: 2
      },
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