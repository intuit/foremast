package ai.foremast.metrics.k8s.starter;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.annotation.Order;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.WebSecurityConfigurerAdapter;

@Configuration
@Order(101)
@EnableConfigurationProperties({K8sMetricsProperties.class})
public class ActuatorSecurityConfig extends WebSecurityConfigurerAdapter {

    @Autowired
    private K8sMetricsProperties metricsProperties;

    /*
        This spring security configuration does the following

        1. Allow access to the all APIs (/info /health /prometheus).
     */
    @Override
    protected void configure(HttpSecurity http) throws Exception {
        if (metricsProperties.isDisableCsrf()) { //For special use case
            http.csrf().disable();
        }
        http.authorizeRequests()
                .antMatchers("/actuator/info", "/actuator/health",
                        "/actuator/prometheus", "/metrics", "/prometheus")
                .permitAll();
    }
}
