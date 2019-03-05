package ai.foremast.metrics.k8s.starter;

import ai.foremast.micrometer.autoconfigure.MeterRegistryCustomizer;
import ai.foremast.micrometer.autoconfigure.MetricsProperties;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ConfigurableApplicationContext;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.regex.Pattern;

/**
 * Auto metrics configurations
 */
@Configuration
public class K8sMetricsAutoConfiguration implements MeterRegistryCustomizer {

    @Autowired
    K8sMetricsProperties k8sMetricsProperties;

    @Autowired
    ConfigurableApplicationContext appContext;

    private static final String HTTP_SERVER_REQUESTS = "http.server.requests";

    @Bean
    public K8sMetricsProperties k8sMetricsProperties() {
        return k8sMetricsProperties = new K8sMetricsProperties();
    }

    @Bean
    public CommonMetricsFilter commonMetricsFilter(K8sMetricsProperties k8sMetricsProperties, MetricsProperties properties) {
        return new CommonMetricsFilter(k8sMetricsProperties, properties);
    }


    @Override
    public void customize(MeterRegistry registry) {
        String commonTagNameValuePair = k8sMetricsProperties.getCommonTagNameValuePairs();
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

        String statuses = k8sMetricsProperties.getInitializeForStatuses();
        if (statuses != null || !statuses.isEmpty()) {
            String[] statusCodes = statuses.split(",");
            for(String code: statusCodes) {
                if (k8sMetricsProperties.hasCaller()) {
                    registry.timer(HTTP_SERVER_REQUESTS, "exception", "None", "method", "GET", "status", code, "uri", "/**", "caller", "*");
                }
                else {
                    registry.timer(HTTP_SERVER_REQUESTS, "exception", "None", "method", "GET", "status", code, "uri", "/**");
                }
            }
        }
    }
}