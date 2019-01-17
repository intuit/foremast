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

import java.util.List;

import io.micrometer.core.annotation.Timed;
import io.micrometer.core.instrument.Clock;

import org.springframework.boot.autoconfigure.AutoConfigureAfter;
import org.springframework.boot.autoconfigure.AutoConfigureBefore;
import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.autoconfigure.amqp.RabbitAutoConfiguration;
import org.springframework.boot.autoconfigure.cache.CacheAutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.jdbc.DataSourceAutoConfiguration;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.annotation.Order;

/**
 * {@link EnableAutoConfiguration} for Micrometer-based metrics.
 *
 * @author Jon Schneider
 */
@Configuration
@ConditionalOnClass(Timed.class)
@EnableConfigurationProperties(MetricsProperties.class)
@AutoConfigureAfter({
        DataSourceAutoConfiguration.class,
        RabbitAutoConfiguration.class,
        CacheAutoConfiguration.class
})
@AutoConfigureBefore(CompositeMeterRegistryAutoConfiguration.class)
public class MetricsAutoConfiguration {
    @Bean
    @ConditionalOnMissingBean(Clock.class)
    public Clock micrometerClock() {
        return Clock.SYSTEM;
    }

    @Bean
    public static MeterRegistryPostProcessor meterRegistryPostProcessor() {
        return new MeterRegistryPostProcessor();
    }

    @Bean
    @Order(0)
    public PropertiesMeterFilter propertiesMeterFilter(MetricsProperties properties) {
        return new PropertiesMeterFilter(properties);
    }

}
