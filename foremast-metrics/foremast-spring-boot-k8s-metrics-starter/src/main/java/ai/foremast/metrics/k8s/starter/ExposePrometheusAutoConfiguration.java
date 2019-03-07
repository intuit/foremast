package ai.foremast.metrics.k8s.starter;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.actuate.autoconfigure.endpoint.web.WebEndpointAutoConfiguration;
import org.springframework.boot.actuate.autoconfigure.endpoint.web.WebEndpointProperties;
import org.springframework.boot.autoconfigure.AutoConfigureBefore;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import javax.annotation.PostConstruct;
import java.util.Set;

@Configuration
@AutoConfigureBefore(WebEndpointAutoConfiguration.class)
@EnableConfigurationProperties({WebEndpointProperties.class})
public class ExposePrometheusAutoConfiguration {


    @Autowired
    WebEndpointProperties webEndpointProperties;

    @PostConstruct
    public void enablePrometheus() {
        if (webEndpointProperties != null) {
            Set<String> set = webEndpointProperties.getExposure().getInclude();
            if (set.isEmpty()) { //Default
                set.add("info");
                set.add("health");
            }
            set.add("prometheus");
            set.add("k8s-metrics");
        }
    }

}
