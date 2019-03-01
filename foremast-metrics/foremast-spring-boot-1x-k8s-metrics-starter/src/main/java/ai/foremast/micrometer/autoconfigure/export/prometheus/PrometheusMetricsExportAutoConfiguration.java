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
package ai.foremast.micrometer.autoconfigure.export.prometheus;

import ai.foremast.micrometer.export.prometheus.PrometheusScrapeMvcEndpoint;
import io.micrometer.core.instrument.Clock;
import io.micrometer.prometheus.PrometheusConfig;
import io.micrometer.prometheus.PrometheusMeterRegistry;
import ai.foremast.micrometer.autoconfigure.CompositeMeterRegistryAutoConfiguration;
import ai.foremast.micrometer.autoconfigure.MetricsAutoConfiguration;
import ai.foremast.micrometer.autoconfigure.export.StringToDurationConverter;
import ai.foremast.micrometer.autoconfigure.export.simple.SimpleMetricsExportAutoConfiguration;
import ai.foremast.micrometer.export.prometheus.PrometheusScrapeEndpoint;
import io.prometheus.client.CollectorRegistry;
import org.springframework.boot.actuate.autoconfigure.ManagementContextConfiguration;
import org.springframework.boot.actuate.condition.ConditionalOnEnabledEndpoint;
import org.springframework.boot.actuate.endpoint.AbstractEndpoint;
import org.springframework.boot.autoconfigure.AutoConfigureAfter;
import org.springframework.boot.autoconfigure.AutoConfigureBefore;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Import;

/**
 * Configuration for exporting metrics to Prometheus.
 *
 * @author Jon Schneider
 * @author David J. M. Karlsen
 * @author Johnny Lim
 */
@Configuration
@AutoConfigureBefore({CompositeMeterRegistryAutoConfiguration.class,
        SimpleMetricsExportAutoConfiguration.class})
@AutoConfigureAfter(MetricsAutoConfiguration.class)
@ConditionalOnBean(Clock.class)
@ConditionalOnClass(PrometheusMeterRegistry.class)
@ConditionalOnProperty(prefix = "management.metrics.export.prometheus", name = "enabled", havingValue = "true", matchIfMissing = true)
@EnableConfigurationProperties(PrometheusProperties.class)
@Import(StringToDurationConverter.class)
public class PrometheusMetricsExportAutoConfiguration {

    @Bean
    @ConditionalOnMissingBean(PrometheusConfig.class)
    public PrometheusConfig prometheusConfig(PrometheusProperties props) {
        return new PrometheusPropertiesConfigAdapter(props);
    }

    @Bean
    @ConditionalOnMissingBean(PrometheusMeterRegistry.class)
    public PrometheusMeterRegistry prometheusMeterRegistry(PrometheusConfig config, CollectorRegistry collectorRegistry,
                                                           Clock clock) {
        return new PrometheusMeterRegistry(config, collectorRegistry, clock);
    }

    @Bean
    @ConditionalOnMissingBean(CollectorRegistry.class)
    public CollectorRegistry collectorRegistry() {
        return new CollectorRegistry(true);
    }

    @ManagementContextConfiguration
    @ConditionalOnClass(AbstractEndpoint.class)
    public static class PrometheusScrapeEndpointConfiguration {
        @Bean
        public PrometheusScrapeEndpoint prometheusEndpoint(CollectorRegistry collectorRegistry) {
            return new PrometheusScrapeEndpoint(collectorRegistry);
        }

        @Bean
        @ConditionalOnEnabledEndpoint("prometheus")
        public PrometheusScrapeMvcEndpoint prometheusMvcEndpoint(PrometheusScrapeEndpoint delegate) {
            return new PrometheusScrapeMvcEndpoint(delegate);
        }
    }

}
