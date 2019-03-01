import _ from 'lodash';

import initialState from '../store/initialState';
import {
  ADD_BASE_METRIC, RECEIVE_METRIC_DATA,
  ADD_ANNOTATION_METRIC, RECEIVE_ANNOTATION_METRIC_DATA
} from '../actions/actionTypes';
import { BASE, UPPER, LOWER, ANOMALY } from '../config/metrics';
import { dataStepValSec } from "../config/api";

export default function metric(state = initialState.metric, action) {
  let newState;
  switch (action.type) {
    case ADD_BASE_METRIC:
      newState = {
        ...state
      };
      newState.resultsByName[action.name] = {
        ...action.object,
        [BASE]: [],
        [UPPER]: [],
        [LOWER]: [],
        [ANOMALY]: {}
      };
      return newState;
    case RECEIVE_METRIC_DATA:
      newState = {
        ...state,
        resultsByName: {
          ...state.resultsByName,
          [action.baseName]: {
            ...state.resultsByName[action.baseName],
            [action.metricType]:
              //each action result resp. provides just 1 data series, so use
              //the 0th index of data arr
              parseMetricData(action.results[0], action.metricType, action.scale,
                state.resultsByName[action.baseName][BASE])
          }
        }
      };
      return newState;
    case ADD_ANNOTATION_METRIC:
      newState = {
        ...state,
        annotationQuery: action.query
      };
      return newState;
    case RECEIVE_ANNOTATION_METRIC_DATA:
      newState = {
        ...state,
        annotations: parseAnnotationData(action.results)
      };
      return newState;
    default:
      return state;
  }
}

const parseMetricData = (data, type, scale, baseSeries) => {
  switch (type) {
    case BASE:
    case UPPER:
    case LOWER:
      //for any direct series, map raw points to plot-able points via timestamp
      //change and converting string value to float, while scaling; empty array
      //if no input data present
      return (data && _.isArray(data.values)) ?
        data.values.map(point => [1000 * point[0],
          scale * parseFloat(point[1])]) : [];
    case ANOMALY:
      return parseAnomalyData(data, baseSeries);
    default:
      break;
  }
};

const parseAnomalyData = (data, baseSeries) => {
  let returnData = [];
  let anomalyArr = (data && _.isArray(data.values)) ?
    data.values.map(point => 1000 * parseInt(point[1])) : [];
  let seen = new Set();
  anomalyArr.forEach(anomalyTimestamp => {
    seen.add(anomalyTimestamp);
  });
  let uniqueAnomalyTimestamps = [];
  for (let time of seen.keys()) {
    //NOTE: this presumes chronological ordering in original data response; not sure if that is warranted or should be explicitly sorted here
    uniqueAnomalyTimestamps.push(time);
  }
  //TODO: DM any way to clean up this n * m processing?
  //for each anomaly timestamp, see if it exists in the base series and use its y value, if so
  uniqueAnomalyTimestamps.forEach(anomalyTimestamp => {
    baseSeries.forEach(basePoint => {
      let timeDiff = anomalyTimestamp - basePoint[0];
      //use this point if it's within a minute (data resolution requested), but only if BEFORE anomaly stamp
      if(timeDiff <= dataStepValSec * 2000 && timeDiff > 0) {
        //NOTE: using base point here will allow for anomalous points to fall directly on top of measured series BUT does therefore indicate slightly different timing than the anaomalies may be marked with
        //NOTE: also, this strategy allows for out of order points to be added, highcharts will warn about this with error #15, but it doesn't stop it from rendering as expected
        returnData.push(basePoint);
      }
    });
  });
  return {
    name: 'Anomaly',
    data: returnData,
    color: '#FF0000',
    marker: {
      enabled: true,
      symbol: 'circle'
    },
    lineWidth: 0,
    states: {
      hover: {
        lineWidthPlus: 0
      }
    }
  };
};

const parseAnnotationData = (data) => {
  //TODO:DM - it feels like some of this could be clean-ed up both in terms of naming and need for all this state
  let activeName = '';
  let newName;
  let annotations = [];
  let min, nextIdcs;
  let allPointsExhausted = false;
  let anyNotExhausted;

  //helper function to be used below; avoids creating function in while loop
  //need to define in parse func to have closure over necessary fn state
  const checkMetricArray = (val, idx) => {
    //compare nextIdc timestamp values to find min
    //ensure our indices don't grow larger than arrs
    if(val.values.length > nextIdcs[idx]){
      //if this is the new min value (earliest time) seen
      if(val.values[nextIdcs[idx]][0] < min.val){
        min.val = val.values[nextIdcs[idx]][0];
        min.idx = idx;
      }
      //multiple at same timestamp
      else if (val.values[nextIdcs[idx]][0] === min.val){
        //increment past this effectively duplicate timestamp
        nextIdcs[idx]++;
      }
      anyNotExhausted = true;
    }
  };

  //only parse to find transition when there are 2 or more arrays in resp
  if(data.length >= 2) {
    //create array of zeros of same length
    nextIdcs = new Array(data.length).fill(0);
    //while not done with arrays O(N) where N is the total number of timesteps
    //across all arrs
    while(!allPointsExhausted){
      min = {val: Infinity, idx: -1};
      anyNotExhausted = false;
      //for each arr
      data.forEach(checkMetricArray);
      //incr that arr's idx
      nextIdcs[min.idx]++;
      //decide active name
      newName = data[min.idx] ? data[min.idx].metric.label_version : '';
      if(newName && activeName !== newName){
        //ensure we're changing from a non-empty name and this time doesn't
        //represent the same as an earlier annotation
        if(activeName){
          annotations.push({timestamp: min.val * 1000, name: newName})
        }
        //eitherway, update name for next iteration
        activeName = newName;
      }
      //update predicates as necessary
      allPointsExhausted = !anyNotExhausted;
    }
  }
  return annotations
};