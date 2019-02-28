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
package ai.foremast.micrometer.controller;


import io.prometheus.client.CollectorRegistry;
import io.prometheus.client.exporter.common.TextFormat;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import org.springframework.http.ResponseEntity;

import java.io.IOException;
import java.io.StringWriter;
import java.io.Writer;

import static org.springframework.http.HttpHeaders.CONTENT_TYPE;


/**
 * Prometheus Controller
 * @author Sheldon Shao
 * @version 1.0
 */
@RestController
@RequestMapping("/actuator")
public class PrometheusController {

    @Autowired
    private CollectorRegistry collectorRegistry;

    @GetMapping(name="/prometheus", produces = "text/plain")
    public ResponseEntity<String> exposeMetrics() {
        try {
            Writer writer = new StringWriter();
            TextFormat.write004(writer, collectorRegistry.metricFamilySamples());
            return ResponseEntity.ok()
                    .header(CONTENT_TYPE, TextFormat.CONTENT_TYPE_004)
                    .body(writer.toString());
        } catch (IOException e) {
            // This actually never happens since StringWriter::write() doesn't throw any IOException
            throw new RuntimeException("Writing metrics failed", e);
        }
    }
}
