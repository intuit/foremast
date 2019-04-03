import { DATA_PATH, DATA_QUERY_PARAM, DATA_START_PARAM,
  DATA_END_PARAM, DATA_STEP_PARAM, DATA_STEP_VAL_SEC } from '../config/api';

export default class ApiService {
  static getMetricData(namespace, appName, metric, startTimestamp, endTimestamp) {
    let tags = tagBuilder(namespace, appName, metric.namespace_key);
    return ApiService.getData(metric.name + tags,
      startTimestamp, endTimestamp);
  }
  static getAnnotationData(query, startTimestamp, endTimestamp) {
    return ApiService.getData(query, startTimestamp, endTimestamp);
  }
  static getData(queryStr, startTimestamp, endTimestamp) {
    let uri = DATA_PATH;
    let params = {
      [DATA_QUERY_PARAM]: queryStr,
      [DATA_START_PARAM]: startTimestamp,
      [DATA_END_PARAM]: endTimestamp,
      [DATA_STEP_PARAM]: DATA_STEP_VAL_SEC
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

//NOTE: string concatenation is potentially insecure IF the string generated
//here were to be directly built into a DB query at any point; but really, this
//sanitization will need to be confirmed in service/back-end layers
const tagBuilder = (namespace = '', appName, namespaceKey) => {
  return `{${namespaceKey}="${namespace}", app="${appName}"}`;
};