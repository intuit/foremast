package ai.foremast.metrics.k8s.starter;


import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

@Component
public class K8sMetricsProperties {

    @Value( "${k8s.metrics.common-tag-name-value-pairs}" )
    private String commonTagNameValuePairs = "app:ENV.APP_NAME|info.app.name";

    @Value("${k8s.metrics.initialize-for-statuses}")
    private String initializeForStatuses = "404,501";

    public static final String APP_ASSET_ALIAS_HEADER = "X-CALLER";

    @Value("${k8s.metrics.caller-header}")
    private String callerHeader = APP_ASSET_ALIAS_HEADER;

    @Value("${k8s.metrics.enable-common-metrics-filter}")
    private boolean enableCommonMetricsFilter = false;

    @Value("${k8s.metrics.common-metrics-whitelist}")
    private String commonMetricsWhitelist = null;

    @Value("${k8s.metrics.common-metrics-blacklist}")
    private String commonMetricsBlacklist = null;

    @Value("${k8s.metrics.common-metrics-prefix}")
    private String commonMetricsPrefix = null;

    @Value("${k8s.metrics.enable-common-metrics-filter-action}")
    private boolean enableCommonMetricsFilterAction = false;


    public String getInitializeForStatuses() {
        return initializeForStatuses;
    }

    public void setInitializeForStatuses(String initializeForStatuses) {
        this.initializeForStatuses = initializeForStatuses;
    }

    public String getCommonTagNameValuePairs() {
        return commonTagNameValuePairs;
    }

    public void setCommonTagNameValuePairs(String commonTagNameValuePairs) {
        this.commonTagNameValuePairs = commonTagNameValuePairs;
    }

    public String getCallerHeader() {
        return callerHeader;
    }

    public boolean hasCaller() {
        return callerHeader != null && !callerHeader.isEmpty();
    }

    public void setCallerHeader(String callerHeader) {
        this.callerHeader = callerHeader;
    }

    public boolean isEnableCommonMetricsFilter() {
        return enableCommonMetricsFilter;
    }

    public void setEnableCommonMetricsFilter(boolean enableCommonMetricsFilter) {
        this.enableCommonMetricsFilter = enableCommonMetricsFilter;
    }

    public String getCommonMetricsWhitelist() {
        return commonMetricsWhitelist;
    }

    public void setCommonMetricsWhitelist(String commonMetricsWhitelist) {
        this.commonMetricsWhitelist = commonMetricsWhitelist;
    }

    public String getCommonMetricsBlacklist() {
        return commonMetricsBlacklist;
    }

    public void setCommonMetricsBlacklist(String commonMetricsBlacklist) {
        this.commonMetricsBlacklist = commonMetricsBlacklist;
    }

    public String getCommonMetricsPrefix() {
        return commonMetricsPrefix;
    }

    public void setCommonMetricsPrefix(String commonMetricsPrefix) {
        this.commonMetricsPrefix = commonMetricsPrefix;
    }

    public boolean isEnableCommonMetricsFilterAction() {
        return enableCommonMetricsFilterAction;
    }

    public void setEnableCommonMetricsFilterAction(boolean enableCommonMetricsFilterAction) {
        this.enableCommonMetricsFilterAction = enableCommonMetricsFilterAction;
    }
}
