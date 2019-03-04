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
package ai.foremast.micrometer.autoconfigure.web.tomcat;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.binder.tomcat.TomcatMetrics;
import ai.foremast.micrometer.web.tomcat.TomcatMetricsBinder;
import org.apache.catalina.Manager;
import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.boot.autoconfigure.condition.ConditionalOnBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnMissingBean;
import org.springframework.boot.autoconfigure.condition.ConditionalOnWebApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * {@link EnableAutoConfiguration Auto-configuration} for {@link TomcatMetrics}.
 *
 * @author Clint Checketts
 * @author Jon Schneider
 * @author Andy Wilkinson
 * @since 1.1.0
 */
@Configuration
@ConditionalOnWebApplication
@ConditionalOnClass({ TomcatMetrics.class, Manager.class })
public class TomcatMetricsAutoConfiguration {

    @Bean
    @ConditionalOnBean(MeterRegistry.class)
    @ConditionalOnMissingBean({ TomcatMetrics.class, TomcatMetricsBinder.class })
    public TomcatMetricsBinder tomcatMetricsBinder(MeterRegistry meterRegistry) {
        return new TomcatMetricsBinder(meterRegistry);
    }

}
