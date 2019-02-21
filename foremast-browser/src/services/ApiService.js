import { dataDomain, dataPath, dataQueryParam, dataStartParam,
  dataEndParam, dataStepParam, dataStepValSec } from '../config/api';

export default class ApiService {
  static getMetricData(metric, startTimestamp, endTimestamp) {
    return ApiService.getData(metric.name + metric.tags,
      startTimestamp, endTimestamp);
  }
  static getAnnotationData(query, startTimestamp, endTimestamp) {
    return ApiService.getData(query, startTimestamp, endTimestamp);
  }
  static getData(queryStr, startTimestamp, endTimestamp) {
    let uri = dataDomain + dataPath;
    let params = {
      [dataQueryParam]: queryStr,
      [dataStartParam]: startTimestamp,
      [dataEndParam]: endTimestamp,
      [dataStepParam]: dataStepValSec
    };
    return requestHelper(uri, params);
  }
}

const requestHelper = (uri, params) => {
  uri += encodeParams(params);
  return fetch(uri)
    .then(res => makeResponse(res))
    .catch(error => {
      throw error;
    });
};

const makeResponse = resp => {
  return new Promise((resolve, reject) => {
    if (resp.ok) {
      resp.json().then(respStr => {
        let parsedResp = JSON.parse(respStr);
        resolve(parsedResp.data.result);
      });
    } else {
      reject('Response object not OK');
    }
  });
};

const encodeParams = params => {
  let paramStr = Object.keys(params).map((key) => {
    return encodeURIComponent(key) + '=' + encodeURIComponent(params[key])
  }).join('&');
  //use falseyness of empty string created by join when params is empty
  if (paramStr){
    paramStr = '?' + paramStr
  }
  return paramStr;
};