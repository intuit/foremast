# Foremast Trigger

## Purpose
Utilize the machine learning capabilities of foremast to monitor services with metrics in wavefront.

## Setup
1. Clone repo into your $GOPATH/src/foremast.ai/foremast
2. Create a file with a semicolon-separated list of services, metrics and corresponding wavefront queries in the form:
  ```
  appname;metric1;metric1query;metric2;metric2query...
  fds-antv;appdynamics.apm.transactions.errors_per_min;sum(align(60s, mean, ts(appdynamics.apm.transactions.errors_per_min, env=prd and app=my-app )), app);appdynamics.apm.transactions.90th_percentile_resp_time_ms;avg(align(60s, mean, ts(appdynamics.apm.transactions.90th_percentile_resp_time_ms, env=prd and app=my-app )), app);appdynamics.apm.transactions.calls_per_min;sum(align(60s, mean, ts(appdynamics.apm.transactions.calls_per_min, env=prd and app=my-app and env=prd)), app)/60
  ```
  the first column is always the Name of the service as it would be found in wavefront and the rest are pairs of appdynamics metric names and the appdynamics queries.
3. Set the necessary environment variables
  ```
  export WAVEFRONT_ENDPOINT=https://your.wavefront.endpoint
  export WAVEFRONT_TOKEN=your-wavefront-auth-token
  export FOREMAST_ES_ENDPOINT=https://your.elasticsearch.endpoint:9200
  export FOREMAST_SERVICE_ENDPOINT=http://your-foremast-service.endpoint:8099
  export VOLUME_PATH=/thedesired/location/for/anomaly/logs/
  export REQUESTS_FILE=/the/file/you/made/earlier.csv
  ```
  or in a kubernetes pod spec with
  ```
  - name: REQUESTS_FILE
    value: /the/file/you/made/earlier.csv
  - name: WAVEFRONT_ENDPOINT
    value: "https://intuit.wavefront.com"
    ...
  ```
4. To run locally, run `go run main.go` from the `foremast/foremast-trigger/cmd/manager` directory while foremast service and elasticsearch are running locally

5. Alternatively, modify the dockerfile at the root to include copying the the service list file made earlier and build an image using `docker build -t something/foremast-trigger` from the foremast-trigger root and deploy

## Function
Foremast Trigger will read the list of applications and the queries from the file and will send an ApplicationHealthAnalyzeRequest to Foremast Service for each application listed. After the Foremast service creates a job in elastic search for the foremast brain to process, the trigger polls the service until the job is completed. Once completed, the trigger sends a new request to continually monitor that service.

When a job is completed and the response to the trigger shows that the application is unhealthy, a line is written to a log in the VOLUME_PATH with the application name, JobID, detailed information about the anomalous data found and a link to the metrics in question in wavefront at the time of the anomaly.
