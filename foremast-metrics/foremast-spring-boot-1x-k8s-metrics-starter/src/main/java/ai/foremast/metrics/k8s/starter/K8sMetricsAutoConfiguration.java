package ai.foremast.metrics.k8s.starter;

import ai.foremast.micrometer.autoconfigure.MeterRegistryCustomizer;
import ai.foremast.micrometer.autoconfigure.MetricsProperties;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.regex.Pattern;

/**
 * Auto metrics configurations
 */
@Configuration
@EnableConfigurationProperties({K8sMetricsProperties.class, MetricsProperties.class})
public class K8sMetricsAutoConfiguration implements MeterRegistryCustomizer {

    @Autowired
    K8sMetricsProperties metricsProperties;

    @Autowired
    ConfigurableApplicationContext appContext;

    private static final String HTTP_SERVER_REQUESTS = "http.server.requests";

    @Bean
    public CommonMetricsFilter commonMetricsFilter(MetricsProperties properties) {
        return new CommonMetricsFilter(metricsProperties, properties);
    }

    @Override
    public void customize(MeterRegistry registry) {
        String commonTagNameValuePair = metricsProperties.getCommonTagNameValuePairs();
        if (commonTagNameValuePair != null && !commonTagNameValuePair.isEmpty()) {
            String[] pairs = commonTagNameValuePair.split(",");
            for(String p : pairs) {
                String[] nameAndValue = p.split(":");
                if (nameAndValue == null || nameAndValue.length != 2) {
                    throw new IllegalStateException("Invalid common tag name value pair:" + p);
                }

                String valuePattern = nameAndValue[1];
                int sepIndex = valuePattern.indexOf('|');
                String[] patterns = null;
                if (sepIndex > 0) {
                    patterns = valuePattern.split(Pattern.quote("|"));
                }
                else {
                    patterns = new String[] { valuePattern };
                }
                for(int i = 0; i < patterns.length; i ++) {
                    String value = null;
                    if (patterns[i].startsWith("ENV.")) {
                        value = System.getenv(patterns[i].substring(4));
                    }
                    else {
                        value = System.getProperty(patterns[i]);
                        if ((value == null || value.isEmpty()) && appContext != null) {
                            value = appContext.getEnvironment().getProperty(patterns[i]);
                        }
                    }

                    if (value != null && !value.isEmpty()) {
                        registry.config().commonTags(nameAndValue[0], value);
                        break;
                    }
                }
            }
        }

        String statuses = metricsProperties.getInitializeForStatuses();
        if (statuses != null || !statuses.isEmpty()) {
            String[] statusCodes = statuses.split(",");
            for(String code: statusCodes) {
                if (metricsProperties.hasCaller()) {
                    registry.timer(HTTP_SERVER_REQUESTS, "exception", "None", "method", "GET", "status", code, "uri", "/**", "caller", "*");
                }
                else {
                    registry.timer(HTTP_SERVER_REQUESTS, "exception", "None", "method", "GET", "status", code, "uri", "/**");
                }
            }
        }
    }
}