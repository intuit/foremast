export const BASE = 'base';
export const UPPER = 'upper';
export const LOWER = 'lower';
export const ANOMALY = 'anomaly';

export const X_METRIC_NAME = 'namespace_app_pod_http_server_requests_latency';
export const Y_METRIC_NAME = 'namespace_app_pod_http_server_requests_errors_5xx';

export const DEFAULT_NAMESPACE = 'dev-container-foremast-examples-usw2-dev-dev';
export const DEFAULT_APPNAME = 'demo';

export const METRICS_MAP = {
  'namespace_app_pod_http_server_requests_errors_5xx': {
    commonName: '5XX Errors',
    metrics: [{
      type: BASE,
      name: 'namespace_app_pod_http_server_requests_errors_5xx',
      namespace_key: 'namespace',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_pod_http_server_requests_errors_5xx_upper',
      namespace_key: 'exported_namespace',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_pod_http_server_requests_errors_5xx_lower',
      namespace_key: 'exported_namespace',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_pod_http_server_requests_errors_5xx_anomaly',
      namespace_key: 'exported_namespace',
    }],
    scale: 1, //multiply by data value
    unit: 'count',
  },
  'namespace_app_pod_http_server_requests_latency': {
    commonName: 'Latency',
    metrics: [{
      type: BASE,
      name: 'namespace_app_pod_http_server_requests_latency',
      namespace_key: 'namespace',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_pod_http_server_requests_latency_upper',
      namespace_key: 'exported_namespace',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_pod_http_server_requests_latency_lower',
      namespace_key: 'exported_namespace',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_pod_http_server_requests_latency_anomaly',
      namespace_key: 'exported_namespace',
    }],
    scale: 1000,
    unit: 'ms',
  },
  'namespace_app_pod_cpu_usage_seconds_total': {
    commonName: 'CPU',
    metrics: [{
      type: BASE,
      name: 'namespace_app_pod_cpu_usage_seconds_total',
      namespace_key: 'namespace',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_pod_cpu_usage_seconds_total_upper',
      namespace_key: 'exported_namespace',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_pod_cpu_usage_seconds_total_lower',
      namespace_key: 'exported_namespace',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_pod_cpu_usage_seconds_total_anomaly',
      namespace_key: 'exported_namespace',
    }],
    scale: 100,
    unit: '%',
  },
  'namespace_app_pod_memory_usage_bytes': {
    commonName: 'Memory',
    metrics: [{
      type: BASE,
      name: 'namespace_app_pod_memory_usage_bytes',
      namespace_key: 'namespace',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_pod_memory_usage_bytes_upper',
      namespace_key: 'exported_namespace',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_pod_memory_usage_bytes_lower',
      namespace_key: 'exported_namespace',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_pod_memory_usage_bytes_anomaly',
      namespace_key: 'exported_namespace',
    }],
    scale: 0.000001,
    unit: 'MB',
  }
};

export const ANNOTATION_QUERY_A =
  'sum by (label_version) (kube_pod_labels{label_app="';
export const ANNOTATION_QUERY_B = '", namespace="';
export const ANNOTATION_QUERY_C = '"})';
