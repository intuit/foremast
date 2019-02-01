package ai.foremast.metrics.k8s.starter;

import org.springframework.boot.actuate.endpoint.annotation.ReadOperation;
import org.springframework.boot.actuate.endpoint.annotation.Selector;
import org.springframework.boot.actuate.endpoint.web.annotation.WebEndpoint;

/**
 *
 *
 * /k8s_metrics/enable/sample_metric_name
 * /k8s_metrics/disable/sample_metric_name
 *
 */
@WebEndpoint(
        id = "k8s_metrics"
)
public class K8sMetricsEndpoint {

    private CommonMetricsFilter commonMetricsFilter;

    @ReadOperation(
            produces = {"text/plain; version=0.1.4; charset=utf-8"}
    )
    public String action(@Selector String action, @Selector String metricName) {
        if (commonMetricsFilter != null) {
            if ("enable".equalsIgnoreCase(action)) {
                commonMetricsFilter.enableMetric(metricName);
            }
            else if ("disable".equalsIgnoreCase(action)) {
                commonMetricsFilter.disableMetric(metricName);
            }
        }
        return "OK";
    }

    public CommonMetricsFilter getCommonMetricsFilter() {
        return commonMetricsFilter;
    }

    public void setCommonMetricsFilter(CommonMetricsFilter commonMetricsFilter) {
        this.commonMetricsFilter = commonMetricsFilter;
    }
}
