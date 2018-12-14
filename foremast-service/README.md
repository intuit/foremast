#foremast-serice
Provide framework to analysis the newly deployed build health.

set up step:

set up elasticsearch start docker docker run -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:6.4.2

get clone current git

go run main.go api example: 1 . create request http://[foremast-service endpoint]:8099/v1/healthcheck/create
body :

{  
   "appName":"k8s-metrics-demo",
   "startTime":"2018-11-06T21:35:30-08:00",
   "endTime":"2018-11-06T21:38:30-08:00",
   "metrics":{  
      "current":{  
         "error4xx":{  
            "dataSourceType":"prometheus",
            "parameters":{  
               "end":1541569110,
               "endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/",
               "query":"namespace_pod:http_server_requests_error_4xx{namespace=\"default\", pod=\"k8s-metrics-demo-7687b9f4d7-k8w6j\"}",
               "start":1541568930,
               "step":60
            }
         },
         "latency":{  
            "dataSourceType":"prometheus",
            "parameters":{  
               "end":1541569110,
               "endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/",
               "query":"namespace_pod:http_server_requests_latency{namespace=\"default\", pod=\"k8s-metrics-demo-7687b9f4d7-k8w6j\"}",
               "start":1541568930,
               "step":60
            }
         }
      },
      "baseline":{  
         "error4xx":{  
            "dataSourceType":"prometheus",
            "parameters":{  
               "end":1541568930,
               "endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/",
               "query":"namespace_pod:http_server_requests_error_4xx{namespace=\"default\", pod=\"k8s-metrics-demo-5db89899b5-lt25r\"}",
               "start":1541568750,
               "step":60
            }
         },
         "latency":{  
            "dataSourceType":"prometheus",
            "parameters":{  
               "end":1541568930,
               "endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/",
               "query":"namespace_pod:http_server_requests_latency{namespace=\"default\", pod=\"k8s-metrics-demo-5db89899b5-lt25r\"}",
               "start":1541568750,
               "step":60
            }
         }
      },
      "historical":{  
         "error4xx":{  
            "dataSourceType":"prometheus",
            "parameters":{  
               "end":1541568930,
               "endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/",
               "query":"namespace_app:http_server_requests_error_4xx{namespace=\"default\", app=\"k8s-metrics-demo\"}",
               "start":1540964130,
               "step":60
            }
         },
         "latency":{  
            "dataSourceType":"prometheus",
            "parameters":{  
               "end":1541568930,
               "endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/",
               "query":"namespace_app:http_server_requests_latency{namespace=\"default\", app=\"k8s-metrics-demo\"}",
               "start":1540964130,
               "step":60
            }
         }
      }
   },
   "strategy":"canary"
}


Response : will return id for example : 08a332dba7ed5c58b148f4409efec50202315e180ec527b0f1b2de205d2fe192 2. search by id http://[foremast-service endpoint]:8099/v1/healthcheck/id/:id for example : http://localhost:8099/v1/healthcheck/id/08a332dba7ed5c58b148f4409efec50202315e180ec527b0f1b2de205d2fe192


You can find entry under 
http://[elastic search host]:9200/documents/_search?pretty=true

kibana :
docker run -d --rm -e "ELASTICSEARCH_URL=http://aa41f5f30e2f011e8bde30674acac93e-1024276836.us-west-2.elb.amazonaws.com:9200"  -p 5601:5601 docker.elastic.co/kibana/kibana:6.0.0
