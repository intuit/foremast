/**
 * Licensed to the Foremast under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package ai.foremast.micrometer.web.servlet;


import ai.foremast.metrics.k8s.starter.CommonMetricsFilter;
import io.prometheus.client.CollectorRegistry;
import io.prometheus.client.exporter.common.TextFormat;


import org.springframework.web.context.WebApplicationContext;
import org.springframework.web.context.support.WebApplicationContextUtils;

import javax.servlet.ServletConfig;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.StringWriter;


/**
 * Prometheus Controller
 * @author Sheldon Shao
 * @version 1.0
 */
public class PrometheusServlet extends HttpServlet {

    private CollectorRegistry collectorRegistry;


    private CommonMetricsFilter commonMetricsFilter;


    public void init(ServletConfig config) throws ServletException {
        super.init(config);

        WebApplicationContext ctx = WebApplicationContextUtils.getWebApplicationContext(config.getServletContext());
        this.setCollectorRegistry(ctx.getBean(CollectorRegistry.class));
        this.setCommonMetricsFilter(ctx.getBean(CommonMetricsFilter.class));

    }


    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws ServletException, IOException {
        String action = req.getParameter("action");
        if (getCommonMetricsFilter() != null && getCommonMetricsFilter().isActionEnabled() && action != null) {
            String metricName = req.getParameter("metric");
            if ("enable".equalsIgnoreCase(action)) {
                getCommonMetricsFilter().enableMetric(metricName);
            }
            else if ("disable".equalsIgnoreCase(action)) {
                getCommonMetricsFilter().disableMetric(metricName);
            }
            resp.getWriter().write("OK");
            return;
        }

        try {
            StringWriter writer = new StringWriter();
            TextFormat.write004(writer, getCollectorRegistry().metricFamilySamples());
            resp.setContentType(TextFormat.CONTENT_TYPE_004);
            resp.getWriter().write(writer.toString());
        } catch (IOException e) {
            // This actually never happens since StringWriter::write() doesn't throw any IOException
            throw new RuntimeException("Writing metrics failed", e);
        }
    }

    public CollectorRegistry getCollectorRegistry() {
        return collectorRegistry;
    }

    public void setCollectorRegistry(CollectorRegistry collectorRegistry) {
        this.collectorRegistry = collectorRegistry;
    }

    public CommonMetricsFilter getCommonMetricsFilter() {
        return commonMetricsFilter;
    }

    public void setCommonMetricsFilter(CommonMetricsFilter commonMetricsFilter) {
        this.commonMetricsFilter = commonMetricsFilter;
    }
}
