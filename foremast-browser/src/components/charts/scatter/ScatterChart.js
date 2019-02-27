import React from 'react';
import Highcharts from 'highcharts';
import Highcharts3d from 'highcharts/highcharts-3d'; //necessary for 'scatter3d' series type
import HighchartsReact from 'highcharts-react-official';

Highcharts3d(Highcharts);

//TODO:DM - how to make this next block of Highcharts event grabbing better fit into react app
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
}, 1000); //1000 ms is totally arbitrary and really a hack to make sure the event binders occur after charts are loaded

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

export default class ScatterChart extends React.Component {
  render() {
    let options = {
      chart: {
        type: 'scatter3d',
        animation: false,
        options3d: {
          enabled: true,
          depth: 400,
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
        text: 'Time by 5XX Errors by Latency'
      },
      xAxis: {
        title: {
          text: 'Time'
        },
        type: 'datetime'
      },
      time: {
        timezoneOffset: 8 * 60, //TODO: CONSTANTIZE me
      },
      yAxis: {
        title: {
          text: 'Error Count'
        }
      },
      zAxis:{
        title: {
          text: 'Latency (ms)'
        }
      },
      tooltip: {
        formatter: function() {
          return  (Highcharts.dateFormat('%e%b%Y %H:%M:%S', new Date(this.x)) +
            '<br/>Error Count: ' + this.y.toFixed(0) +
            '<br/>Latency:' + this.point.z.toFixed(2) + ' ms');
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
    if(xSeries.length !== 0 && ySeries.length !== 0){
      xSeries.length = Math.min(xSeries.length, ySeries.length);
      ySeries.length = Math.min(xSeries.length, ySeries.length);
    }
    //NOTE: this causes us to either LOSE data (from longer series) or create data (for shorter series)
    //series lengths may differ as they are loaded asynchronously, only build series once they are the same length
    if (xSeries.length === ySeries.length) {
      //TODO:DM - seems fragile to just presume that the points in both series have same sequence of timestamps
      for(let i = 0; i < xSeries.length; i++){
        xytData.push([
          ySeries[i][0],
          ySeries[i][1],
          xSeries[i][1],
        ]);
      }
    }
    let xytSeries = {
      type: 'scatter3d',
      name: 'Errors by Latency',
      data: xytData,
      marker: {
        radius: 2
      },
    };
    return [xytSeries];
  }
}