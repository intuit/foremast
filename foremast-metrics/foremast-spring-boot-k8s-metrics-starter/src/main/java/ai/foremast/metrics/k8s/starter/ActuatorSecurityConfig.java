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
    private K8sMetricsProperties k8sMetricsProperties;

    /*
        This spring security configuration does the following

        1. Allow access to specific APIs (/info /health /prometheus).
        2. Allow access to "/actuator/k8s-metrics/*"
     */
    @Override
    protected void configure(HttpSecurity http) throws Exception {
        if (k8sMetricsProperties.isDisableCsrf()) {
            http.csrf().disable();
        }
        http.authorizeRequests()
                .antMatchers("/actuator/info", "/actuator/health", "/actuator/prometheus", "/metrics", "/actuator/k8s-metrics/*")
                .permitAll();
    }
}
