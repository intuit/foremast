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
package ai.foremast.metrics.k8s.starter;

import io.micrometer.core.instrument.Meter;
import io.micrometer.core.instrument.config.MeterFilter;
import io.micrometer.core.instrument.config.MeterFilterReply;
import org.springframework.boot.actuate.autoconfigure.metrics.MetricsProperties;
import org.springframework.util.StringUtils;

import java.util.*;
import java.util.function.Supplier;

/**
 * Common Metrics Filter is to hide the metrics by default, and show the metrics what are in the whitelist.
 *
 * If the metric is enabled in properties, keep it enabled.
 * If the metric is disabled in properties, keep it disabled.
 * If the metric starts with "PREFIX", enable it. otherwise disable the metric
 *
 * @author Sheldon Shao
 * @version 1.0
 */
public class CommonMetricsFilter implements MeterFilter {

    private MetricsProperties properties;

    private String[] prefixes = new String[] {};

    private LinkedHashMap<String, String> tagRules = new LinkedHashMap<>();

    private Set<String> whitelist = new HashSet<>();

    private Set<String> blacklist = new HashSet<>();

    private K8sMetricsProperties k8sMetricsProperties;

    public CommonMetricsFilter(K8sMetricsProperties k8sMetricsProperties, MetricsProperties properties) {
        this.properties = properties;
        this.k8sMetricsProperties = k8sMetricsProperties;
        String list = k8sMetricsProperties.getCommonMetricsBlacklist();
        if (list != null && !list.isEmpty()) {
            String[] array = StringUtils.tokenizeToStringArray(list, ",", true, true);
            for(String str: array) {
                str = filter(str.trim());
                blacklist.add(str);
            }
        }
        list = k8sMetricsProperties.getCommonMetricsWhitelist();
        if (list != null && !list.isEmpty()) {
            String[] array = StringUtils.tokenizeToStringArray(list, ",", true, true);
            for(String str: array) {
                str = filter(str.trim());
                whitelist.add(str);
            }
        }

        list = k8sMetricsProperties.getCommonMetricsPrefix();
        if (list != null && !list.isEmpty()) {
            prefixes = StringUtils.tokenizeToStringArray(list, ",", true, true);
        }

        String tagRuleExpressions = k8sMetricsProperties.getCommonMetricsTagRules();
        if (tagRuleExpressions != null) {
            String[] array = StringUtils.tokenizeToStringArray(tagRuleExpressions, ",", true, true);
            for(String str: array) {
                str = str.trim();
                String[] nameAndValue = str.split(":");
                if (nameAndValue == null || nameAndValue.length != 2) {
                    throw new IllegalStateException("Invalid common tag name value pair:" + str);
                }
                tagRules.put(nameAndValue[0].trim(), nameAndValue[1].trim());
            }
        }

    }

    /**
     * Filter for "[PREFIX]_*"
     *
     * @param id
     * @return MeterFilterReply
     */
    public MeterFilterReply accept(Meter.Id id) {
        if (!k8sMetricsProperties.isEnableCommonMetricsFilter()) {
            return MeterFilter.super.accept(id);
        }

        String metricName = id.getName();

        Boolean enabled = lookupWithFallbackToAll(this.properties.getEnable(), id, null);
        if (enabled != null) {
            return enabled ? MeterFilterReply.NEUTRAL : MeterFilterReply.DENY;
        }

        if (whitelist.contains(metricName)) {
            return MeterFilterReply.NEUTRAL;
        }
        if (blacklist.contains(metricName)) {
            return MeterFilterReply.DENY;
        }

        for(String prefix: prefixes) {
            if (metricName.startsWith(prefix)) {
                return MeterFilterReply.ACCEPT;
            }
        }

        for(String key: tagRules.keySet()) {
            String expectedValue = tagRules.get(key);
            if (expectedValue != null) {
                if (expectedValue.equals(id.getTag(key))) {
                    return MeterFilterReply.ACCEPT;
                }
            }
        }

        return MeterFilterReply.DENY;
    }

    protected String filter(String metricName) {
        return metricName.replace('_', '.');
    }

    public void enableMetric(String metricName) {
        metricName = filter(metricName);

        if (blacklist.contains(metricName)) {
            blacklist.remove(metricName);
        }
        if (!whitelist.contains(metricName)) {
            whitelist.add(metricName);
        }
    }

    public void disableMetric(String metricName) {
        metricName = filter(metricName);

        if (whitelist.contains(metricName)) {
            whitelist.remove(metricName);
        }
        if (!blacklist.contains(metricName)) {
            blacklist.add(metricName);
        }
    }

    private <T> T lookup(Map<String, T> values, Meter.Id id, T defaultValue) {
        if (values.isEmpty()) {
            return defaultValue;
        }
        return doLookup(values, id, () -> defaultValue);
    }

    private <T> T lookupWithFallbackToAll(Map<String, T> values, Meter.Id id, T defaultValue) {
        if (values.isEmpty()) {
            return defaultValue;
        }
        return doLookup(values, id, () -> values.getOrDefault("all", defaultValue));
    }

    private <T> T doLookup(Map<String, T> values, Meter.Id id, Supplier<T> defaultValue) {
        String name = id.getName();
        while (name != null && !name.isEmpty()) {
            T result = values.get(name);
            if (result != null) {
                return result;
            }
            int lastDot = name.lastIndexOf('.');
            name = (lastDot != -1) ? name.substring(0, lastDot) : "";
        }

        return defaultValue.get();
    }

    public String[] getPrefixes() {
        return prefixes;
    }

    public void setPrefixes(String[] prefixes) {
        this.prefixes = prefixes;
    }
}
