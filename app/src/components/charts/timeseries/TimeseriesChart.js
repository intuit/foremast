import React from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import PropTypes from 'prop-types';
import Highcharts from 'highcharts';
import HighchartsMore from 'highcharts/highcharts-more'; //necessary for 'arearange' series type
import HighchartsReact from 'highcharts-react-official';

import * as highlightActions from '../../../actions/highlightActions';

HighchartsMore(Highcharts);

(function(H) {
  H.Pointer.prototype.reset = function() {
    return undefined;
  };

  /**
   * Highlight a point by showing tooltip, setting hover state and draw crosshair
   */
  H.Point.prototype.highlight = function(event) {
    event = this.series.chart.pointer.normalize(event);
    this.onMouseOver(); // Show the hover marker
    //this.series.chart.tooltip.refresh(this); // Show the tooltip
    this.series.chart.xAxis[0].drawCrosshair(event, this); // Show the crosshair
  };

  H.syncExtremes = function(e) {
    var thisChart = this.chart;

    if (e.trigger !== "syncExtremes") {
      // Prevent feedback loop
      Highcharts.each(Highcharts.charts, function(chart) {
        if (chart && chart !== thisChart) {
          if (chart.xAxis[0].setExtremes) {
            // It is null while updating
            chart.xAxis[0].setExtremes(e.min, e.max, undefined, false, {
              trigger: "syncExtremes"
            });
          }
        }
      });
    }
  };
})(Highcharts);

class TimeseriesChart extends React.Component {
  render() {
    const timeseriesOptions = {
      chart: {
        zoomType: 'x',
        type: 'line',
        height: 180,
      },
      title: {
        text: this.props.metricName,// + ' Metric + Modeled Range',
        style: {"fontSize": "12px"}
      },
      // subtitle: {
      //   text: document.ontouchstart === undefined ?
      //     'Click and drag in the plot area to zoom in' : 'Pinch the chart to zoom in',
      //   style: {"fontSize": "8px"}
      // },
      xAxis: {
        type: 'datetime',
        events: {
          setExtremes: function(e) {
            Highcharts.syncExtremes(e);
          }
        }
      },
      time: {
        timezoneOffset: 8 * 60, //TODO: CONSTANTIZE me //8hrs offset, will need to be changed on DST shifts
      },
      yAxis: {
        title: {
          //TODO:DM - change this based on metric name selected, may need to hard-code unit with rest of data in App.js
          text: this.props.unit,
        },
        min: 0,
      },
      series: this.buildSeries(),
      tooltip: {
        valueDecimals: 5,
        split: true,
        distance: 30,
        padding: 5
      },
      legend: {
        enabled: false
      },
      navigator: {
        enabled: false
      },
      plotOptions: {
        line: {
          marker: {
            enabled: false
          }
        },
        arearange: {
          marker: {
            enabled: false
          }
        },
        series: {
          label: {
            connectorAllowed: false
          },
          point: {
            events: {
              mouseOver: function () {
                //highlightActions.updateHighlightTimestamp(this.x);
              }
            }
          },
          events: {
            mouseOut: function () {
              //highlightActions.clearHighlightTimestamp();
            }
          }
        }
      },

      responsive: {
        rules: [{
          condition: {
            maxWidth: 500
          },
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
        }, {
          type: 'all',
          text: 'all',
        }],
      },
      credits: {
        enabled: false
      },
    };

    return (
      <HighchartsReact
        highcharts={Highcharts}
        options={timeseriesOptions}
      />
    );
  }

  buildSeries() {
    let rangeData = [];
    const { baseSeries, upperSeries, lowerSeries, anomalySeries } = this.props;
    //series lengths may differ as they are loaded asynchronously, only build series once they are the same length
    if (upperSeries.data.length === lowerSeries.data.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < upperSeries.data.length; i++){
        rangeData.push([
          upperSeries.data[i][0],
          lowerSeries.data[i][1],
          upperSeries.data[i][1]
        ]);
      }
    }
    let rangeSeries = {
      type: 'arearange',
      showInLegend: false,
      name: 'Model Range',
      data: rangeData,
      fillOpacity: 0.1,
      color: '#A6EA8A'
    };
    return [baseSeries, anomalySeries, rangeSeries];
  }
}

TimeseriesChart.propTypes = {
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
)(TimeseriesChart);