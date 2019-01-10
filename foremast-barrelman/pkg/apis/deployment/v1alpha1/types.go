package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=DeploymentMetadatas
// +domain=foremast.ai

// DeploymentMetadata describes a DeploymentMetadata.
type DeploymentMetadata struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DeploymentMetadataSpec `json:"spec"`
	// The status object for the Application.
	Status DeploymentMetadataStatus `json:"status,omitempty"`
}

// DeploymentMetadataSpec is the spec for a DeploymentMetadata resource
type DeploymentMetadataSpec struct {
	Analyst Analyst `json:"analyst"`

	// Description is human readable content explaining the purpose of the link.
	Description string `json:"description,omitempty"`

	// Metrics metadata
	Metrics Metrics `json:"metrics"`

	// Logs should be monitored by canary deployment
	Logs []Logs `json:"logs,omitempty"`

	// Descriptor regroups information and metadata about an application.
	Descriptor Descriptor `json:"descriptor,omitempty"`
}

// DeploymentMetadataStatus defines controllers the observed state of DeploymentMetadata
type DeploymentMetadataStatus struct {
	// ObservedGeneration is used by the DeploymentMetadata Controller to report the last Generation of a DeploymentMetadata
	// that it has observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// Analyst backend information

type Analyst struct {
	Endpoint string `json:"endpoint"`

	//Default is 0.0.1
	Version string `json:"version,omitempty"`
}

// The following descriptor
// https://github.com/kubernetes-sigs/application/blob/master/pkg/apis/app/v1beta1/application_types.go
// Descriptor defines the Metadata and information about the DeploymentMetadata.
type Descriptor struct {
	// Type is the type of the application (e.g. SpringBoot, NodeJS, Python).
	Type string `json:"type,omitempty"`

	// Version is an optional version indicator for the Application.
	Version string `json:"version,omitempty"`

	// Description is a brief string description of the Application.
	Description string `json:"description,omitempty"`

	// Icons is an optional list of icons for an application. Icon information includes the source, size,
	// and mime type.
	Icons []ImageSpec `json:"icons,omitempty"`

	// Maintainers is an optional list of maintainers of the application. The maintainers in this list maintain the
	// the source code, images, and package for the application.
	Maintainers []ContactData `json:"maintainers,omitempty"`

	// Owners is an optional list of the owners of the installed application. The owners of the application should be
	// contacted in the event of a planned or unplanned disruption affecting the application.
	Owners []ContactData `json:"owners,omitempty"`

	// Keywords is an optional list of key words associated with the application (e.g. MySQL, RDBMS, database).
	Keywords []string `json:"keywords,omitempty"`

	// Links are a list of descriptive URLs intended to be used to surface additional documentation, dashboards, etc.
	Links []Link `json:"links,omitempty"`

	// Notes contain a human readable snippets intended as a quick start for the users of the Application.
	// CommonMark markdown syntax may be used for rich text representation.
	Notes string `json:"notes,omitempty"`
}

// ImageSpec contains information about an image used as an icon.
type ImageSpec struct {
	// The source for image represented as either an absolute URL to the image or a Data URL containing
	// the image. Data URLs are defined in RFC 2397.
	Source string `json:"src"`

	// (optional) The size of the image in pixels (e.g., 25x25).
	Size string `json:"size,omitempty"`

	// (optional) The mine type of the image (e.g., "image/png").
	Type string `json:"type,omitempty"`
}

// ContactData contains information about an individual or organization.
type ContactData struct {
	// Name is the descriptive name.
	Name string `json:"name,omitempty"`

	// Url could typically be a website address.
	Url string `json:"url,omitempty"`

	// Email is the email address.
	Email string `json:"email,omitempty"`
}

// Link contains information about an URL to surface documentation, dashboards, etc.
type Link struct {
	// Description is human readable content explaining the purpose of the link.
	Description string `json:"description,omitempty"`

	// Url typically points at a website address.
	Url string `json:"url,omitempty"`
}

// Metrics metadata
type Metrics struct {

	// prometheus or other metrics solution
	DataSourceType string `json:"dataSourceType"`

	// Endpoint of prometheus
	Endpoint string `json:"endpoint"`

	// Monitoring contains the metrics should be monitored by canary deployment
	Monitoring []Monitoring `json:"monitoring,omitempty"`
}

type Monitoring struct {
	MetricName string `json:"metricName"`

	// Gauge counter timer etc. Default is counter
	MetricType string `json:"metricType,omitempty"`

	// Shorten name in backend system
	MetricAlias string `json:"metricAlias"`
}

type Logs struct {
	LogName string `json:"logName"`

	LogType string `json:"logType"`

	// Default is "logName" + ".log"
	FilePattern string `json:"filePattern,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeploymentMetadataList is a list of DeploymentMetadata resources
type DeploymentMetadataList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeploymentMetadata `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=DeploymentMonitors
// +groupName=deployment.foremast.ai

// DeploymentMonitor describes a DeploymentMonitor information.
type DeploymentMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DeploymentMonitorSpec `json:"spec"`
	// The status object for the Application.
	Status DeploymentMonitorStatus `json:"status,omitempty"`
}

// DeploymentMonitorSpec is the spec for a DeploymentMonitor resource

//  Actions could be remediation action AND triggering alerts
//  Since remediation action should be triggered very carefully. We are going to support only one action this time.
//  Ideally system should cover all the use cases automatically. so we should only need a SUPER_SMART remediation action in the future.
//  But alerts could be different ways, such as sending email or sending slack message, either way should take care of duplication messages,
//  We'd like to integrate with existing smart open source solutions which support those use cases very well in the future,
//  We will send foremast internal metrics so that we can define AlertRules in prometheus to generate Alerts

type DeploymentMonitorSpec struct {
	// Selector is a label query over kinds that created by the application. It must match the component objects' labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	Analyst Analyst `json:"analyst,omitempty"`

	StartTime string `json:"startTime,omitempty"`

	WaitUntil string `json:"waitUntil,omitempty"`

	// Metrics metadata
	Metrics Metrics `json:"metrics,omitempty"`

	// Logs should be monitored by canary deployment
	Logs []Logs `json:"logs,omitempty"`

	AutoRollback bool `json:"autoRollback,omitempty"`

	Remediation RemediationAction `json:"remediation,omitempty"`

	// Rollback revision
	RollbackRevision int64 `json:"rollbackRevision,omitempty"`
}

// DeploymentMonitorStatus defines controllers the observed state of DeploymentMonitor
type DeploymentMonitorStatus struct {
	// ObservedGeneration is used by the DeploymentMetadata Controller to report the last Generation of a DeploymentMetadata
	// that it has observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	JobId string `json:"jobId,omitempty"`

	Phase string `json:"phase"`

	RemediationTaken bool `json:"remediationTaken"`

	Anomaly Anomaly `json:"anomaly,omitempty"`

	Timestamp string `json:"timestamp"`

	Expired bool `json:"expired"`
}

// "healthy" deployment is considered healthy, "running" the deployment is running and it is detecting,
// "error" got error, "anomaly" engine detected anomaly issues,
const (
	MonitorPhaseHealthy = "Healthy"

	MonitorPhaseRunning = "Running"

	MonitorPhaseFailed = "Failed"

	MonitorPhaseUnhealthy = "Unhealthy"

	MonitorPhaseWarning = "Warning"

	MonitorPhaseExpired = "Expired"

	MonitorPhaseAbort = "Abort"
)

//Constants for Actions
const (
	//No remediation required
	RemediationNone = "None"
	//Trigger a rollback if error occured
	RemediationAutoRollback = "AutoRollback"
	//Trigger a pause only to reduce the error rate
	RemediationAutoPause = "AutoPause"
	//Trigger an auto scaling for specific use cases, for example, connection stack or CPU bump up a lot
	RemediationAutoScaling = "AutoScaling"
	//Let foremast take care everything for you
	RemediationAuto = "Auto"
)

// Option could be RemediationNone, RemediationAutoRollback, RemediationAutoPause, RemediationAutoScaling, RemediationAuto
//
type RemediationAction struct {
	Option string `json:"option"`

	Parameters map[string]string `json:"parameters,omitempty"`
}

// Anomaly detected
type Anomaly struct {
	AnomalousMetrics []AnomalousMetric `json:"anomalousMetrics,omitempty"`
}

type AnomalousMetric struct {
	Name string `json:"name"`

	Tags string `json:"tags,omitempty"`

	Values []AnomalousMetricValue `json:"values"`
}

type AnomalousMetricValue struct {
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DeploymentMonitorList is a list of DeploymentMonitor resources
type DeploymentMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DeploymentMonitor `json:"items"`
}
