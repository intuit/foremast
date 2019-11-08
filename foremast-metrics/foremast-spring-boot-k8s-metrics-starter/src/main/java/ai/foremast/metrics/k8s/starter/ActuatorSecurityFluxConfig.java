package ai.foremast.metrics.k8s.starter;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.autoconfigure.condition.ConditionalOnClass;
import org.springframework.boot.autoconfigure.condition.ConditionalOnWebApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.core.annotation.Order;
import org.springframework.security.authentication.AuthenticationManager;
import org.springframework.security.config.web.server.ServerHttpSecurity;
import org.springframework.security.web.server.SecurityWebFilterChain;
import org.springframework.security.web.server.util.matcher.ServerWebExchangeMatchers;

@Configuration
@Order(101)
@EnableConfigurationProperties({K8sMetricsProperties.class})
@ConditionalOnClass(AuthenticationManager.class)
@ConditionalOnWebApplication(type = ConditionalOnWebApplication.Type.REACTIVE)
public class ActuatorSecurityFluxConfig {


    @Autowired
    private K8sMetricsProperties k8sMetricsProperties;

    /*
        This spring security configuration does the following

        1. Allow access to specific APIs (/info /health /prometheus).
        2. Allow access to "/actuator/k8s-metrics/*"
     */

    @Bean
    public SecurityWebFilterChain apiHttpSecurity(
            ServerHttpSecurity http) {
        if (k8sMetricsProperties.isDisableCsrf()) {
            http.csrf().disable();
        }
        http.securityMatcher(ServerWebExchangeMatchers.pathMatchers("/actuator/info", "/actuator/health", "/actuator/prometheus", "/metrics", "/actuator/k8s-metrics/*"))
                .authorizeExchange().anyExchange().permitAll();
        return http.build();
    }

}
