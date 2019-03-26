package ai.foremast.metrics.k8s.starter;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "k8s.metrics")
public class K8sMetricsProperties {

    private String commonTagNameValuePairs = "app:ENV.APP_NAME|info.app.name";

    private String initializeForStatuses = "403,404,501,502";

    public static final String APP_ASSET_ALIAS_HEADER = "X-CALLER";

    private String callerHeader = APP_ASSET_ALIAS_HEADER;

    private boolean disableCsrf = false;

    private boolean enableCommonMetricsFilter = false;

    private String commonMetricsWhitelist = null;

    private String commonMetricsBlacklist = null;

    private String commonMetricsPrefix = null;

    private boolean enableCommonMetricsFilterAction = false;

    private String commonMetricsTagRules = null;


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

    public boolean isDisableCsrf() {
        return disableCsrf;
    }

    public void setDisableCsrf(boolean disableCsrf) {
        this.disableCsrf = disableCsrf;
    }

    public String getCommonMetricsTagRules() {
        return commonMetricsTagRules;
    }

    public void setCommonMetricsTagRules(String commonMetricsTagRules) {
        this.commonMetricsTagRules = commonMetricsTagRules;
    }
}
