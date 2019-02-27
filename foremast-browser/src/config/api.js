//domain is either a specific URI for localhost requests, or an empty string
//such that nginx will proxy API requests in other deploy scenarios
export const dataDomain =
  (/localhost/.test(window.location.href)) ?
    'http://foremast-service:8099' : '';
export const dataPath = '/api/v1/query_range';
export const dataQueryParam = 'query';
export const dataStartParam = 'start';
export const dataEndParam = 'end';
export const dataStepParam = 'step';
export const dataStepValSec = 15; //data granularity