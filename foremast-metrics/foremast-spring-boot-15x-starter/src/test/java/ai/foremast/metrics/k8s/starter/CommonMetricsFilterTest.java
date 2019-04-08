package ai.foremast.metrics.k8s.starter;
import io.micrometer.core.instrument.Meter;
import io.micrometer.core.instrument.Tags;
import io.micrometer.core.instrument.config.MeterFilterReply;
import io.micrometer.spring.autoconfigure.MetricsProperties;
import org.junit.Test;
import org.springframework.boot.actuate.metrics.Metric;

import static org.junit.Assert.*;

public class CommonMetricsFilterTest {

    @Test
    public void acceptDisabled() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        filter.setPrefixes(new String[]{"prefix"});
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);


        filter.setPrefixes(new String[]{"prefix"});
        reply = filter.accept(new Meter.Id("abc.something", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);
    }

    @Test
    public void acceptEnabled() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);

        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        filter.setPrefixes(new String[]{"prefix"});
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.ACCEPT);


        filter.setPrefixes(new String[]{"prefix"});
        reply = filter.accept(new Meter.Id("abc_.omething", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.DENY);
    }

    @Test
    public void acceptBlackList() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);
        k8sMetricsProperties.setCommonMetricsBlacklist("prefix_abc");

        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        filter.setPrefixes(new String[]{"prefix"});
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.DENY);  //Backlist
    }

    @Test
    public void acceptWhiteList() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);
        k8sMetricsProperties.setCommonMetricsWhitelist("prefix_abc");

        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);  //Whitelist
    }

    @Test
    public void tagRules() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);
        k8sMetricsProperties.setCommonMetricsTagRules("myTag:true");

        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", Tags.of("myTag", "true"), "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.ACCEPT);

        reply = filter.accept(new Meter.Id("prefix.abc", Tags.of("myTag", "false"), "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.DENY);
    }

    @Test
    public void filter() {
    }

    @Test
    public void enableMetric() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);
        k8sMetricsProperties.setCommonMetricsBlacklist("prefix_abc");
        k8sMetricsProperties.setEnableCommonMetricsFilterAction(true);

        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        filter.setPrefixes(new String[]{"prefix"});
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.DENY);  //Backlist

        filter.enableMetric("prefix_abc");
        reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);
    }

    @Test
    public void disableAction() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);
        k8sMetricsProperties.setCommonMetricsWhitelist("prefix_abc");


        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);  //Whitelist

        filter.disableMetric("prefix_abc");
        reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);
    }

    @Test
    public void disableMetric() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);
        k8sMetricsProperties.setCommonMetricsWhitelist("prefix_abc");
        k8sMetricsProperties.setEnableCommonMetricsFilterAction(true);

        MetricsProperties metricsProperties = new MetricsProperties();
        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.NEUTRAL);  //Whitelist

        filter.disableMetric("prefix_abc");
        reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.DENY);
    }

    @Test
    public void defaultMetricsProperties() {
        K8sMetricsProperties k8sMetricsProperties = new K8sMetricsProperties();
        k8sMetricsProperties.setEnableCommonMetricsFilter(true);

        MetricsProperties metricsProperties = new MetricsProperties();
        metricsProperties.getEnable().put("prefix.abc", false);

        CommonMetricsFilter filter = new CommonMetricsFilter(k8sMetricsProperties, metricsProperties);
        MeterFilterReply reply = filter.accept(new Meter.Id("prefix.abc", null, "", "", Meter.Type.COUNTER));
        assertEquals(reply, MeterFilterReply.DENY);  //Disabled
    }

    @Test
    public void getPrefixes() {
    }

    @Test
    public void setPrefixes() {
    }

    @Test
    public void isActionEnabled() {
    }
}