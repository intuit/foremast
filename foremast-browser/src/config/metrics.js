export const BASE = 'base';
export const UPPER = 'upper';
export const LOWER = 'lower';
export const ANOMALY = 'anomaly';

export const METRICS_MAP = {
  'namespace_app_per_pod:http_server_requests_error_5xx': {
    commonName: '5XX Errors',
    metrics: [{
      type: BASE,
      name: 'namespace_app_per_pod:http_server_requests_error_5xx',
      tags: '{namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_per_pod:http_server_requests_error_5xx_upper',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_per_pod:http_server_requests_error_5xx_lower',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_per_pod:http_server_requests_error_5xx_anomaly',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }],
    scale: 1, //multiply by data value
    unit: 'count',
  },
  'namespace_app_per_pod:http_server_requests_latency': {
    commonName: 'Latency',
    metrics: [{
      type: BASE,
      name: 'namespace_app_per_pod:http_server_requests_latency',
      tags: '{namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_per_pod:http_server_requests_latency_upper',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_per_pod:http_server_requests_latency_lower',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_per_pod:http_server_requests_latency_anomaly',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }],
    scale: 1000,
    unit: 'ms',
  },
  'namespace_app_per_pod:cpu_usage_seconds_total': {
    commonName: 'CPU',
    metrics: [{
      type: BASE,
      name: 'namespace_app_per_pod:cpu_usage_seconds_total',
      tags: '{namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_upper',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_lower',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_per_pod:cpu_usage_seconds_total_anomaly',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }],
    scale: 100,
    unit: '%',
  },
  'namespace_app_per_pod:memory_usage_bytes': {
    commonName: 'Memory',
    metrics: [{
      type: BASE,
      name: 'namespace_app_per_pod:memory_usage_bytes',
      tags: '{namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: UPPER,
      name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_upper',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: LOWER,
      name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_lower',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }, {
      type: ANOMALY,
      name: 'foremastbrain:namespace_app_per_pod:memory_usage_bytes_anomaly',
      tags: '{exported_namespace="dev-container-foremast-examples-usw2-dev-dev",app="demo"}',
    }],
    scale: 0.000001,
    unit: 'MB',
  }
};

export const ANNOTATION_QUERY =
  'sum by (label_version) (kube_pod_labels{label_app="demo", namespace="dev-container-foremast-examples-usw2-dev-dev"})';