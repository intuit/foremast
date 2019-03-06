import React from 'react';
import Highcharts from 'highcharts';
import HighchartsMore from 'highcharts/highcharts-more'; //necessary for 'arearange' series type
import HighchartsReact from 'highcharts-react-official';

HighchartsMore(Highcharts);

//TODO:DM - how to better encapsulate this in a react friendly way?
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

export default class TimeseriesChart extends React.Component {
  render() {
    const timeseriesOptions = {
      chart: {
        zoomType: 'x',
        type: 'line',
        height: 180,
      },
      title: {
        text: this.props.metricName,
        style: {"fontSize": "12px"}
      },
      xAxis: {
        type: 'datetime',
        plotLines: this.buildAnnotations(),
        crosshair: {
          enabled: true
        },
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
      rangeSelector:{
        enabled:false
      },
      scrollbar: {
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
          pointInterval: 15 * 1000,
          label: {
            connectorAllowed: false
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
    if (upperSeries.length === lowerSeries.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < upperSeries.length; i++){
        rangeData.push([
          upperSeries[i][0],
          lowerSeries[i][1],
          upperSeries[i][1]
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
    return [
      {name: 'Measured', data: baseSeries},
      anomalySeries, rangeSeries
    ];
  }

  buildAnnotations() {
    const { annotations } = this.props;
    let plotLines = [];
    annotations.forEach(annotation => {
      plotLines.push({
        color: 'grey',
        width: 2,
        value: annotation.timestamp,
        label: {
          text: annotation.name
        }
      });
    });
    return plotLines;
  }
}