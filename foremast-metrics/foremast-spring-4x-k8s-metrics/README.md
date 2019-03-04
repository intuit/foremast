# Exposes metrics for spring 4.x applications + tomcat


#### Features
1. Expose http_server_requests related metrics
2. Expose JVM gc related metrics.
3. Expose Tomcat metrics
4. Support customized metrics

#### Properties
```properties
management.security.enabled=false

management.metrics.use-global-registry=false

management.metrics.export.prometheus.enabled=true

endpoints.prometheus.id=prometheus
endpoints.prometheus.sensitive=false
endpoints.prometheus.enabled=true


management.metrics.export.prometheus.descriptions=false


# Common tag names, it should follow this syntax  "tagName1:applicationPropertyToGetValue1,tagName2:applicationPropertyToGetValue2"
# For example, we are going to add region as a tag,  the "tagName" is "region"
# The region information can be retrieved in applicationProperties with name "k8s.region" (depends on the real property name)
# Multiple tags could use ',' to separate
k8s.metrics.common-tag-name-value-pairs=app:ENV.APP_NAME|info.app.name,assetId:ENV.ASSET_ID,env:ENV.APP_ENV,l1:ENV.L1,l2:ENV.L2

# Set the following statuses' counter value to 0, so that we can get zero instead of null value from prometheus
k8s.metrics.initialize-for-statuses=403,404,500,503

# Add 90% 95% for http.server.requests
management.metrics.distribution.percentiles.http.server.requests=0.95,0.98

# Caller header, the identity for caller, if set it to empty, then the caller tag will be ignored
k8s.metrics.caller-header=intuit_assetalias


# Use selected common metrics
# https://docs.google.com/document/d/10GMZxAwNR4twrNh0nwmTeAj7TqqqMd7VDbFA-5GagJI/edit?usp=sharing
k8s.metrics.enable-common-metrics-filter=true


# http_server_requests
#
# tomcat_threads_busy        #Total thread in tomcat
# tomcat_threads_current    #Current thread count in tomcat
# tomcat_sessions_active_current #Current active session count
# process_files_open            #How many file handles are opening, to detect TooManyFilesOpenError or socket leak
# process_files_max             #The maximum file descriptor count
# jvm_threads_live                #The current number of live threads including both daemon and non-daemon threads
# jvm_gc_live_data_size_bytes      #Size of old generation memory pool after a full GC
# jvm_gc_pause_seconds_count   #Count of GC pause
# jvm_gc_pause_seconds_sum     #Sum of time spent in GC pause
# jvm_gc_pause_seconds_max     #Max of time spent in GC pause within 2 minutes
# jvm_memory_used_bytes           #The amount of used memory
# system_cpu_usage                     #The recent cpu usage for whole OS(retrieve by JVM)
k8s.metrics.common-metrics-whitelist=http_servers_request,tomcat_threads_busy,tomcat_threads_current,tomcat_sessions_active_current,\
  process_files_open,process_files_max,jvm_threads_live,jvm.gc.live.data.size,jvm_gc_live_data_size_bytes,jvm.gc.pause,jvm_gc_pause_seconds,\
  jvm_memory_used_bytes,jvm.memory.used,system_cpu_usage

# Expose all metrics with intuit prefix
k8s.metrics.common-metrics-prefix=intuit
```

#### How to use in your spring boot application?
```xml
    <properties>
      <foremast-spring-4x-k8s-metrics>0.1.6-SNAPSHOT</foremast-spring-4x-k8s-metrics>
    </properties>

    <dependencies>
    <!-- K8s Metrics Starter which is to simplify the metrics usage on K8s -->
      <dependency>
         <groupId>ai.foremast.metrics</groupId>
         <artifactId>foremast-spring-4x-k8s-metrics</artifactId>
         <version>${foremast-spring-4x-k8s-metrics}</version>
      </dependency>
    </dependencies>
```

```xml
	<context-param>
		<param-name>contextConfigLocation</param-name>
		<param-value>
			classpath:foremast-spring-4x-k8s-metrics.xml
		</param-value>
	</context-param>
	
	<!-- ... -->
    <filter>
        <filter-name>HttpServerMetrics</filter-name>
        <filter-class>ai.foremast.micrometer.web.servlet.WebMvcMetricsFilter</filter-class>
        <init-param>
            <param-name>prometheus-path</param-name>
            <param-value>/actuator/prometheus</param-value>
        </init-param>
    </filter>

    <filter-mapping>
        <filter-name>HttpServerMetrics</filter-name>
        <url-pattern>/*</url-pattern>
    </filter-mapping>
```