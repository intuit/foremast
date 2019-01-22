package ai.foremast.metrics.k8s.starter;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "k8s.metrics")
public class K8sMetricsProperties {

    private String commonTagNameValuePairs = "app:ENV.APP_NAME|info.app.name";

    private String initializeForStatuses = "403,404,501,502";

    public static final String APP_ASSET_ALIAS_HEADER = "X-CALLER";

    private String callerHeader = APP_ASSET_ALIAS_HEADER;

    private boolean disableCsrf = false;

    private boolean disableSecurityConfig = false;

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

    public boolean isDisableSecurityConfig() {
        return disableSecurityConfig;
    }

    public void setDisableSecurityConfig(boolean disableSecurityConfig) {
        this.disableSecurityConfig = disableSecurityConfig;
    }
}
