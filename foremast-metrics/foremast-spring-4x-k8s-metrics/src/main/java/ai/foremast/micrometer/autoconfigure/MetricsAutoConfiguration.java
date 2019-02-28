/**
* Copyright 2017 Pivotal Software, Inc.
* <p>
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* <p>
* http://www.apache.org/licenses/LICENSE-2.0
* <p>
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/
package ai.foremast.micrometer.autoconfigure;

import ai.foremast.micrometer.autoconfigure.export.prometheus.PrometheusProperties;
import ai.foremast.micrometer.autoconfigure.export.prometheus.PrometheusPropertiesConfigAdapter;
import ai.foremast.micrometer.web.servlet.DefaultWebMvcTagsProvider;
import ai.foremast.micrometer.web.servlet.HandlerMappingIntrospector;
import ai.foremast.micrometer.web.servlet.WebMvcMetricsFilter;
import ai.foremast.micrometer.web.servlet.WebMvcTagsProvider;
import ai.foremast.micrometer.web.tomcat.TomcatMetricsBinder;
import io.micrometer.core.instrument.Clock;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.binder.jvm.ClassLoaderMetrics;
import io.micrometer.core.instrument.binder.jvm.JvmGcMetrics;
import io.micrometer.core.instrument.binder.jvm.JvmMemoryMetrics;
import io.micrometer.core.instrument.binder.jvm.JvmThreadMetrics;
import io.micrometer.core.instrument.binder.system.FileDescriptorMetrics;
import io.micrometer.core.instrument.binder.system.ProcessorMetrics;
import io.micrometer.core.instrument.binder.system.UptimeMetrics;
import io.micrometer.core.instrument.composite.CompositeMeterRegistry;
import io.micrometer.core.instrument.config.MeterFilter;
import io.micrometer.prometheus.PrometheusConfig;
import io.micrometer.prometheus.PrometheusMeterRegistry;
import io.prometheus.client.CollectorRegistry;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;
import org.springframework.core.annotation.Order;
import org.springframework.web.context.WebApplicationContext;

import java.util.List;

/**
 *
 * @author Jon Schneider
 */
@Configuration
public class MetricsAutoConfiguration {
    @Bean
    public Clock micrometerClock() {
        return Clock.SYSTEM;
    }

    @Bean
    public static MeterRegistryPostProcessor meterRegistryPostProcessor() {
        return new MeterRegistryPostProcessor();
    }


    @Autowired
    private MetricsProperties properties;


    @Bean
    public MetricsProperties initMetricsProperties() {
        return properties = new MetricsProperties();
    }

    @Autowired
    private PrometheusProperties prometheusProperties;

    @Bean
    public PrometheusProperties initPrometheusProperties() {
        return prometheusProperties = new PrometheusProperties();
    }

    @Bean
    @Order(0)
    public PropertiesMeterFilter propertiesMeterFilter() {
        return new PropertiesMeterFilter(properties);
    }


    @Bean
    public PrometheusConfig prometheusConfig() {
        return new PrometheusPropertiesConfigAdapter(prometheusProperties);
    }

    @Bean
    public PrometheusMeterRegistry prometheusMeterRegistry(PrometheusConfig config, CollectorRegistry collectorRegistry,
                                                           Clock clock) {
        return new PrometheusMeterRegistry(config, collectorRegistry, clock);
    }

    @Bean
    public CollectorRegistry collectorRegistry() {
        return new CollectorRegistry(true);
    }



    @Bean
    public DefaultWebMvcTagsProvider servletTagsProvider() {
        return new DefaultWebMvcTagsProvider();
    }

    @SuppressWarnings("deprecation")
    @Bean
    public WebMvcMetricsFilter webMetricsFilter(MeterRegistry registry,
                                                WebMvcTagsProvider tagsProvider,
                                                WebApplicationContext ctx) {
        return new WebMvcMetricsFilter(registry, tagsProvider,
                properties.getWeb().getServer().getRequestsMetricName(),
                properties.getWeb().getServer().isAutoTimeRequests(),
                new HandlerMappingIntrospector(ctx));
    }

    @Bean
    @Order(0)
    public MeterFilter metricsHttpServerUriTagFilter() {
        String metricName = this.properties.getWeb().getServer().getRequestsMetricName();
        MeterFilter filter = new OnlyOnceLoggingDenyMeterFilter(() -> String
                .format("Reached the maximum number of URI tags for '%s'.", metricName));
        return MeterFilter.maximumAllowableTags(metricName, "uri",
                this.properties.getWeb().getServer().getMaxUriTags(), filter);
    }

    @Bean
    public CompositeMeterRegistry noOpMeterRegistry(Clock clock) {
        return new CompositeMeterRegistry(clock);
    }

    @Bean
    @Primary
    public CompositeMeterRegistry compositeMeterRegistry(Clock clock, List<MeterRegistry> registries) {
        return new CompositeMeterRegistry(clock, registries);
    }

    @Bean
    public TomcatMetricsBinder tomcatMetricsBinder(MeterRegistry meterRegistry) {
        return new TomcatMetricsBinder(meterRegistry);
    }

    @Bean
//    @ConditionalOnProperty(value = "management.metrics.binders.uptime.enabled", matchIfMissing = true)
//    @ConditionalOnMissingBean(UptimeMetrics.class)
    public UptimeMetrics uptimeMetrics() {
        return new UptimeMetrics();
    }

    @Bean
//    @ConditionalOnProperty(value = "management.metrics.binders.processor.enabled", matchIfMissing = true)
//    @ConditionalOnMissingBean(ProcessorMetrics.class)
    public ProcessorMetrics processorMetrics() {
        return new ProcessorMetrics();
    }

    @Bean
//    @ConditionalOnProperty(name = "management.metrics.binders.files.enabled", matchIfMissing = true)
//    @ConditionalOnMissingBean(FileDescriptorMetrics.class)
    public FileDescriptorMetrics fileDescriptorMetrics() {
        return new FileDescriptorMetrics();
    }


    //JVM
    @Bean
    public JvmGcMetrics jvmGcMetrics() {
        return new JvmGcMetrics();
    }

    @Bean
    public JvmMemoryMetrics jvmMemoryMetrics() {
        return new JvmMemoryMetrics();
    }

    @Bean
    public JvmThreadMetrics jvmThreadMetrics() {
        return new JvmThreadMetrics();
    }

    @Bean
    public ClassLoaderMetrics classLoaderMetrics() {
        return new ClassLoaderMetrics();
    }

}
