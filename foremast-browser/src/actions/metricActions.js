import * as types from './actionTypes';
import ApiService from '../services/ApiService';

export const addBaseMetric = (name, object) => ({
  type: types.ADD_BASE_METRIC,
  name,
  object
});

export const requestMetricData = (baseName, metric, scale, start, end) => {
  return dispatch => {
    let requestPromise = ApiService.getMetricData(metric, start, end);
    return requestPromise.then(results => {
      dispatch(receiveMetricData(baseName, metric, scale, results));
    });
  };
};

export const receiveMetricData = (baseName, metric, scale, results) => ({
  type: types.RECEIVE_METRIC_DATA,
  baseName,
  metricType: metric.type,
  scale,
  results
});

export const addAnnotationMetric = query => ({
  type: types.ADD_ANNOTATION_METRIC,
  query
});

export const requestAnnotationMetricData = (query, start, end) => {
  return dispatch => {
    let requestPromise = ApiService.getAnnotationData(query, start, end);
    return requestPromise.then(results => {
      dispatch(receiveAnnotationMetricData(results, query));
    });
  };
};

export const receiveAnnotationMetricData = (results, query) => ({
  type: types.RECEIVE_ANNOTATION_METRIC_DATA,
  results,
  query
});