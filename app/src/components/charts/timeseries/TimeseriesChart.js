import React from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import PropTypes from 'prop-types';
import HighchartsMore from 'highcharts/highcharts-more'; //necessary for 'arearange' series type
import Highstock from 'highcharts/highstock';
import HighchartsReact from 'highcharts-react-official';

import * as highlightActions from '../../../actions/highlightActions';

HighchartsMore(Highstock);

class TimeseriesChart extends React.Component {
  render() {
    const { highlightActions } = this.props;
    const timeseriesOptions = {
      chart: {
        zoomType: 'x'
      },
      title: {
        text: this.props.metricName + ' Metric + Modeled Range',
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
      series: this.buildSeries(),
      tooltip: {
        valueDecimals: 5
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
          point: {
            events: {
              mouseOver: function () {
                console.log(this.x);
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
        highcharts={Highstock}
        constructorType={'stockChart'}
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