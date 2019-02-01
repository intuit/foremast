package ai.foremast.metrics.k8s.starter;

import javax.servlet.*;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

/**
 * K8s use /metrics and prometheus format metrics description by default
 * So "/metrics" needs to be pointed to "/actuator/prometheus"
 *
 *
 * /metrics/enable/sample_metric_name
 * /metrics/disable/sample_metric_name
 *
 */
public class K8sMetricsFilter implements Filter {

    private CommonMetricsFilter commonMetricsFilter;

    @Override
    public void init(FilterConfig filterConfig) {
    }

    @Override
    public void doFilter(ServletRequest servletRequest, ServletResponse servletResponse, FilterChain filterChain) throws IOException, ServletException {
        if (servletResponse instanceof HttpServletResponse) {
            HttpServletRequest request = (HttpServletRequest)servletRequest;
            HttpServletResponse response = (HttpServletResponse)servletResponse;

            String requestPath = request.getServletPath();
            if ("/metrics".equals(requestPath)) {
                response.sendRedirect("/actuator/prometheus");
            }
            else if (requestPath.startsWith("/metrics/enable/") || requestPath.startsWith("/metrics/disable/")) {
                boolean enable = requestPath.startsWith("/metrics/enable/");
                String metricName = requestPath.substring(enable ? "/metrics/enable/".length() : "/metrics/disable/".length());
                if (commonMetricsFilter != null) {
                    if (enable) {
                        commonMetricsFilter.enableMetric(metricName);
                    }
                    else {
                        commonMetricsFilter.disableMetric(metricName);
                    }
                }
                response.getWriter().println("OK");
            }
        }
        else {
            filterChain.doFilter(servletRequest, servletResponse);
        }
    }

    @Override
    public void destroy() {

    }

    public CommonMetricsFilter getCommonMetricsFilter() {
        return commonMetricsFilter;
    }

    public void setCommonMetricsFilter(CommonMetricsFilter commonMetricsFilter) {
        this.commonMetricsFilter = commonMetricsFilter;
    }
}
