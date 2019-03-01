import React from 'react';
import { connect } from "react-redux";
import { bindActionCreators } from "redux";
import { withRouter } from 'react-router-dom';
import SplitterLayout from 'react-splitter-layout';
import moment from 'moment';
import Highcharts from 'highcharts';

import './App.css';
import * as metricActions from "./actions/metricActions";
import { METRICS_MAP, ANNOTATION_QUERY_A, ANNOTATION_QUERY_B,
  ANNOTATION_QUERY_C, BASE, UPPER, LOWER, ANOMALY, X_METRIC_NAME,
  Y_METRIC_NAME, DEFAULT_NAMESPACE, DEFAULT_APPNAME } from './config/metrics';
import { dataStepValSec } from './config/api';
import Header from './components/header/Header';
import TimeseriesChart from './components/charts/timeseries/TimeseriesChart';
import ScatterChart from './components/charts/scatter/ScatterChart';

class App extends React.Component {
  state = {
    namespace: '',
    appName: ''
  };

  componentDidMount() {
    //TODO:DM - can I clean up this HC event gymnastics, make more react?
    ["mousemove", "touchmove", "touchstart"].forEach(function (eventType) {
      document
        .getElementById("container")
        .addEventListener(eventType, function (e) {
          var chart, point, i, event;

          for (i = 0; i < Highcharts.charts.length; i = i + 1) {
            chart = Highcharts.charts[i];
            if (chart) {
              // Find coordinates within the chart
              event = chart.pointer.normalize(e);
              // Get the hovered point
              point = chart.series[0].searchPoint(event, true);
              if (point) {
                point.highlight(e);
              }
            }
          }
        });
    });

    //this next listener allows for tooltips to be hidden once no longer
    //mousing over points
    document.getElementById("container")
      .addEventListener('mouseleave', function (e) {
        for (let j = 0; j < Highcharts.charts.length; j++) {
            let chart = Highcharts.charts[j];
            if (chart) {
              chart.tooltip.hide();
              chart.xAxis[0].hideCrosshair();
            }
          }
      });

    //TODO:DM - figure out how to make this more DRY wrt same block below
    let { namespace, appName } = this.props.match.params;
    //override these params if undefined
    namespace = namespace || DEFAULT_NAMESPACE;
    appName = appName || DEFAULT_APPNAME;

    //track base metrics and annotations based on config
    Object.keys(METRICS_MAP).forEach(key => {
      this.props.metricActions.addBaseMetric(key, METRICS_MAP[key]);
    });

    this.props.metricActions.addAnnotationMetric(ANNOTATION_QUERY_A + appName +
      ANNOTATION_QUERY_B + namespace + ANNOTATION_QUERY_C);

    this.fetchData();
    setInterval(this.fetchData, dataStepValSec * 1000);
  }

  fetchData = () => {
    //API can't provide more than roughly 7 days of data at 60sec granularity,
    //in practice we'll probably request even quite a bit less; 15min here
    let endTimestamp = moment().subtract(0, 'minutes').unix();
    //align requests with 15 sec boundaries to make responses more consistent
    endTimestamp -= (endTimestamp % 15);
    let startTimestamp = moment().subtract(15, 'minutes').unix();
    startTimestamp -= (startTimestamp % 15);

    let { namespace, appName } = this.props.match.params;
    //override these params if undefined
    namespace = namespace || DEFAULT_NAMESPACE;
    appName = appName || DEFAULT_APPNAME;
    this.setState({
      namespace,
      appName
    });

    const { metric, metricActions } = this.props;

    //request data for each metric being tracked and for each of the sub-series
    //(upper, lower, etc.)
    Object.keys(metric.resultsByName).forEach(key => {
      metric.resultsByName[key].metrics.forEach(metricObj => {
        let scale = metric.resultsByName[key].scale;
        metricActions.requestMetricData(namespace, appName, key, metricObj, scale, startTimestamp, endTimestamp);
      });
    });

    //and request updated annotation data
    if(metric.annotationQuery){
     metricActions.requestAnnotationMetricData(metric.annotationQuery, startTimestamp, endTimestamp);
    }
  };

  render() {
    let { namespace, appName } = this.state;
    let { metric } = this.props;
    return (
      <div className="App">
        <Header/>
        <SplitterLayout vertical={true}>
          <div id="container">
            {
              Object.keys(metric.resultsByName).map(key => {
                return (
                  <TimeseriesChart
                    key={key}
                    metricName={namespace + ' : ' + appName + ' : ' +
                      metric.resultsByName[key].commonName}
                    unit={metric.resultsByName[key].unit}
                    baseSeries={metric.resultsByName[key][BASE]}
                    upperSeries={metric.resultsByName[key][UPPER]}
                    lowerSeries={metric.resultsByName[key][LOWER]}
                    anomalySeries={metric.resultsByName[key][ANOMALY]}
                    annotations={metric.annotations}
                  />
                );
              })
            }
          </div>
          <ScatterChart
            xSeries={metric.resultsByName[X_METRIC_NAME] ?
              [...metric.resultsByName[X_METRIC_NAME][BASE]] : []}
            ySeries={metric.resultsByName[Y_METRIC_NAME] ?
              [...metric.resultsByName[Y_METRIC_NAME][BASE]] : []}
          />
        </SplitterLayout>
      </div>
    );
  }
}

const mapStoreToProps = store => ({metric: store.metric});

const mapDispatchToProps = dispatch => ({
  metricActions: bindActionCreators(metricActions, dispatch)
});

export default connect(
  mapStoreToProps,
  mapDispatchToProps
)(withRouter(App));