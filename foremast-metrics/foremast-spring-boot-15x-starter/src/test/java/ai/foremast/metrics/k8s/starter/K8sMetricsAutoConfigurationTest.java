package ai.foremast.metrics.k8s.starter;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.composite.CompositeMeterRegistry;
import org.junit.Test;

import static org.junit.Assert.*;

public class K8sMetricsAutoConfigurationTest {

    @Test
    public void commonMetricsFilter() {
    }

    @Test
    public void customize() {

        MeterRegistry registry = new CompositeMeterRegistry();

        K8sMetricsAutoConfiguration k8sMetricsAutoConfiguration = new K8sMetricsAutoConfiguration();
        k8sMetricsAutoConfiguration.metricsProperties = new K8sMetricsProperties();
        k8sMetricsAutoConfiguration.customize(registry);

    }
}