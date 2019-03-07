# foremast-spring-boot-1.x-k8s-metrics-starter
Metrics and custom metrics setup for spring boot 1.x application on K8s

#### Features

1. Common tags configuration support
2. Added "app" as common tags
3. Default distribution and zero values for http.server.requests
4. URL filter for "/metrics" which points "/metrics" to "/actuator/prometheus", so that K8s can discover "/metrics" by default
5. Default security configuration for actuator
6. Add caller as a tag on http_server_request (Since 0.1.1)
7. Allow to read app name from application.properties

#### Properties
```properties
# Common tag names, it should follow this syntax  "tagName1:applicationPropertyToGetValue1,tagName2:applicationPropertyToGetValue2"
# For example, we are going to add region as a tag,  the "tagName" is "region"
# The region information can be retrieved in applicationProperties with name "k8s.region" (depends on the real property name)
# Multiple tags could use ',' to separate
k8s.metrics.common-tag-name-value-pairs=app:ENV.APP_NAME|info.app.name

# Set the following statuses' counter value to 0, so that we can get zero instead of null value from prometheus
k8s.metrics.initialize-for-statuses=403,404,500,503

# Add 90% 95% for http.server.requests
management.metrics.distribution.percentiles.http.server.requests=0.95,0.98

# Caller header, the identity for caller, if set it to empty, then the caller tag will be ignored
k8s.metrics.callerHeader=X-CALLER
```

#### How to use in your spring boot application?
```xml
    <properties>
      <foremast-spring-boot-1x-k8s-metrics-starter>0.1.6</foremast-spring-boot-1x-k8s-metrics-starter>
    </properties>

    <dependencies>
    <!-- K8s Metrics Starter which is to simplify the metrics usage on K8s -->
      <dependency>
         <groupId>ai.foremast.metrics</groupId>
         <artifactId>foremast-spring-boot-1x-k8s-metrics-starter</artifactId>
         <version>${foremast-spring-boot-1x-k8s-metrics-starter}</version>
      </dependency>
    </dependencies>
```
