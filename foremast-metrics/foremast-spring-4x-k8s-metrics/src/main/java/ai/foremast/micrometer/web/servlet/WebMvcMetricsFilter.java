/**
 * Copyright 2017 Pivotal Software, Inc.
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
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

import ai.foremast.micrometer.autoconfigure.MetricsProperties;
import io.micrometer.core.annotation.Timed;
import io.micrometer.core.instrument.LongTaskTimer;
import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import ai.foremast.micrometer.TimedUtils;
import io.micrometer.core.instrument.binder.tomcat.TomcatMetrics;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;
import org.springframework.http.HttpStatus;
import org.springframework.web.context.WebApplicationContext;
import org.springframework.web.context.support.WebApplicationContextUtils;
import org.springframework.web.method.HandlerMethod;
import org.springframework.web.servlet.DispatcherServlet;
import org.springframework.web.servlet.HandlerExecutionChain;
import org.springframework.web.util.NestedServletException;


import javax.servlet.*;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.Collection;
import java.util.Collections;
import java.util.Enumeration;
import java.util.Set;
import java.util.stream.Collectors;

/**
 * Intercepts incoming HTTP requests and records metrics about execution time and results.
 *
 * @author Jon Schneider
 */
public class WebMvcMetricsFilter implements Filter {
    private static final String TIMING_SAMPLE = "micrometer.timingSample";

    private static final Log logger = LogFactory.getLog(WebMvcMetricsFilter.class);

    private MeterRegistry registry;
    private MetricsProperties metricsProperties;
    private WebMvcTagsProvider tagsProvider = new DefaultWebMvcTagsProvider();
    private String metricName;
    private boolean autoTimeRequests;
    private HandlerMappingIntrospector mappingIntrospector;

    private boolean exposePrometheus = false;
    private String prometheusPath = "/actuator/prometheus";

    private PrometheusServlet prometheusServlet;

    private void record(TimingSampleContext timingContext, HttpServletResponse response, HttpServletRequest request,
                        Object handlerObject, Throwable e) {
        for (Timed timedAnnotation : timingContext.timedAnnotations) {
            timingContext.timerSample.stop(Timer.builder(timedAnnotation, metricName)
                .tags(tagsProvider.httpRequestTags(request, response, handlerObject, e))
                .register(registry));
        }

        if (timingContext.timedAnnotations.isEmpty() && autoTimeRequests) {
            timingContext.timerSample.stop(Timer.builder(metricName)
                .tags(tagsProvider.httpRequestTags(request, response, handlerObject, e))
                .register(registry));
        }

        for (LongTaskTimer.Sample sample : timingContext.longTaskTimerSamples) {
            sample.stop();
        }
    }

    @Override
    public void init(final FilterConfig filterConfig) throws ServletException {
        WebApplicationContext ctx = WebApplicationContextUtils.getWebApplicationContext(filterConfig.getServletContext());

        this.registry = ctx.getBean(MeterRegistry.class);
        this.metricsProperties = ctx.getBean(MetricsProperties.class);
        this.metricName = metricsProperties.getWeb().getServer().getRequestsMetricName();
        this.autoTimeRequests = metricsProperties.getWeb().getServer().isAutoTimeRequests();
        this.mappingIntrospector = new HandlerMappingIntrospector(ctx);

        String path = filterConfig.getInitParameter("prometheus-path");
        if (path != null) {
            exposePrometheus = true;
            this.prometheusPath = path;
        }

        if (exposePrometheus) {
            prometheusServlet = new PrometheusServlet();
            prometheusServlet.init(new ServletConfig() {
                @Override
                public String getServletName() {
                    return "prometheus";
                }

                @Override
                public ServletContext getServletContext() {
                    return filterConfig.getServletContext();
                }

                @Override
                public String getInitParameter(String s) {
                    return filterConfig.getInitParameter(s);
                }

                @Override
                public Enumeration getInitParameterNames() {
                    return filterConfig.getInitParameterNames();
                }
            });
        }

        //Tomcat
        try {
            new TomcatMetrics(null, Collections.emptyList()).bindTo(this.registry);
        }
        catch(Throwable tx) {

        }
    }

    @Override
    public void doFilter(ServletRequest servletRequest, ServletResponse servletResponse, FilterChain filterChain)
            throws IOException, ServletException {
        HttpServletRequest request = (HttpServletRequest)servletRequest;
        HttpServletResponse response = (HttpServletResponse)servletResponse;

        String path = request.getServletPath();
        if (exposePrometheus && prometheusPath.equals(path)) { // Handle /actuator/prometheus URL in this filter
            prometheusServlet.doGet(request, response);
            return;
        }

        HandlerExecutionChain handler = null;
        try {
            handler = mappingIntrospector.getHandlerExecutionChain(request);
        } catch (Exception e) {
            logger.debug("Unable to time request", e);
            filterChain.doFilter(request, response);
            return;
        }

        final Object handlerObject = handler == null ? null : handler.getHandler();

        // If this is the second invocation of the filter in an async request, we don't
        // want to start sampling again (effectively bumping the active count on any long task timers).
        // Rather, we'll just use the sampling context we started on the first invocation.
        TimingSampleContext timingContext = (TimingSampleContext) request.getAttribute(TIMING_SAMPLE);
        if (timingContext == null) {
            timingContext = new TimingSampleContext(request, handlerObject);
        }

        MetricsServletResponse wrapper = new MetricsServletResponse(response);

        try {
            filterChain.doFilter(request, wrapper);

            record(timingContext, wrapper, request,
                    handlerObject, (Throwable) request.getAttribute(EXCEPTION_ATTRIBUTE));
        } catch (NestedServletException e) {
            response.setStatus(HttpStatus.INTERNAL_SERVER_ERROR.value());
            record(timingContext, response, request, handlerObject, e.getCause());
            throw e;
        }
    }

    public static final String EXCEPTION_ATTRIBUTE = DispatcherServlet.class.getName() + ".EXCEPTION";


    @Override
    public void destroy() {

    }

    private class TimingSampleContext {
        private final Set<Timed> timedAnnotations;
        private final Timer.Sample timerSample;
        private final Collection<LongTaskTimer.Sample> longTaskTimerSamples;

        TimingSampleContext(HttpServletRequest request, Object handlerObject) {
            timedAnnotations = annotations(handlerObject);
            timerSample = Timer.start(registry);
            longTaskTimerSamples = timedAnnotations.stream()
                .filter(Timed::longTask)
                .map(t -> LongTaskTimer.builder(t)
                    .tags(tagsProvider.httpLongRequestTags(request, handlerObject))
                    .register(registry)
                    .start())
                .collect(Collectors.toList());
        }

        private Set<Timed> annotations(Object handler) {
            if (handler instanceof HandlerMethod) {
                HandlerMethod handlerMethod = (HandlerMethod) handler;
                Set<Timed> timed = TimedUtils.findTimedAnnotations(handlerMethod.getMethod());
                if (timed.isEmpty()) {
                    return TimedUtils.findTimedAnnotations(handlerMethod.getBeanType());
                }
                return timed;
            }
            return Collections.emptySet();
        }
    }
}
