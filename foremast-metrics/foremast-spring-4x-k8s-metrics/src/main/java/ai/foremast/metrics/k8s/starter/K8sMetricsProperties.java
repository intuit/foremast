package ai.foremast.metrics.k8s.starter;


import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

//@ConfigurationProperties(prefix = "k8s.metrics")
@Component
public class K8sMetricsProperties {

    @Value( "${k8s.metrics.common-tag-name-value-pairs}" )
    private String commonTagNameValuePairs = "app:ENV.APP_NAME|info.app.name";

    @Value("${k8s.metrics.initialize-for-statuses}")
    private String initializeForStatuses = "403,404,501,502";

    public static final String APP_ASSET_ALIAS_HEADER = "X-CALLER";

    @Value("${k8s.metrics.callerHeader}")
    private String callerHeader = APP_ASSET_ALIAS_HEADER;


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
}
