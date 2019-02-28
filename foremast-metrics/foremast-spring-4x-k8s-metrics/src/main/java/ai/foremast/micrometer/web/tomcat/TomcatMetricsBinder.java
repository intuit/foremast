/**
 * Copyright 2019 Pivotal Software, Inc.
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
package ai.foremast.micrometer.web.tomcat;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Tag;
import io.micrometer.core.instrument.binder.tomcat.TomcatMetrics;
import org.apache.catalina.core.ApplicationContextFacade;
import org.springframework.context.event.ContextStartedEvent;

import javax.servlet.ServletContainerInitializer;
import javax.servlet.ServletContext;
import javax.servlet.ServletException;
import java.util.Collections;
import java.util.Set;

/**
 * Binds {@link TomcatMetrics} in response to the {@link ContextStartedEvent}.
 *
 * @author Andy Wilkinson
 * @since 1.1.0
 */
public class TomcatMetricsBinder implements ServletContainerInitializer {

    private final MeterRegistry meterRegistry;

    private final Iterable<Tag> tags;

    public TomcatMetricsBinder(MeterRegistry meterRegistry) {
        this(meterRegistry, Collections.emptyList());
    }

    public TomcatMetricsBinder(MeterRegistry meterRegistry, Iterable<Tag> tags) {
        this.meterRegistry = meterRegistry;
        this.tags = tags;
    }

//    @Override
//    public void onApplicationEvent(ContextStartedEvent event) {
//        ApplicationContext applicationContext = event.getApplicationContext();
//        Manager manager = findManager(applicationContext);
//        new TomcatMetrics(manager, this.tags).bindTo(this.meterRegistry);
//    }
//
//    private Manager findManager(ApplicationContext applicationContext) {
//        applicationContext.executeMethod()
//        if (applicationContext instanceof ApplicationContextFacade) {
//            EmbeddedServletContainer container = ((EmbeddedWebApplicationContext) applicationContext).getEmbeddedServletContainer();
//            if (container instanceof TomcatEmbeddedServletContainer) {
//                Context context = findContext((TomcatEmbeddedServletContainer) container);
//                if (context != null) {
//                    return context.getManager();
//                }
//            }
//        }
//        return null;
//    }
//
//    private Context findContext(TomcatEmbeddedServletContainer tomcatWebServer) {
//        for (Container container : tomcatWebServer.getTomcat().getHost().findChildren()) {
//            if (container instanceof Context) {
//                return (Context) container;
//            }
//        }
//        return null;
//    }

    @Override
    public void onStartup(Set<Class<?>> set, ServletContext servletContext) throws ServletException {
        ApplicationContextFacade facade = null;
        if (servletContext instanceof ApplicationContextFacade) {
            facade = (ApplicationContextFacade)servletContext;
        }
        else if (servletContext instanceof org.apache.catalina.core.ApplicationContext) {
            facade = new ApplicationContextFacade((org.apache.catalina.core.ApplicationContext)servletContext);
        }
        else {
            throw new IllegalStateException("The servlet container is not tomcat");
        }

    }
}
