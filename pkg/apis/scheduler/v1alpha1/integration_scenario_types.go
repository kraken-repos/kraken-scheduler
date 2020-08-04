package v1alpha1

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// IntegrationScenario is the Schema for the integrationscenarios API.
// +k8s:openapi-gen=true
type IntegrationScenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IntegrationScenarioSpec   `json:"spec,omitempty"`
	Status IntegrationScenarioStatus `json:"status,omitempty"`
}

// Check that IntegrationScenario can be validated and can be defaulted.
var _ runtime.Object = (*IntegrationScenario)(nil)
var _ kmeta.OwnerRefable = (*IntegrationScenario)(nil)

type DomainExtractionFilterStrategySpec struct {
	// IsFilterEnabled is the flag to enable Client-side filtering strategy in Integration Agent.
	// +required
	IsFilterEnabled string `json:"isFilterEnabled"`

	// FilterByField is the field to be used for Client-side filtering strategy in Integration Agent.
	// +optional
	FilterByField string `json:"filterByField,omitempty"`

	// FilterByFieldExpectedValue is the expected value of the field to be used for Client-side filtering strategy in Integration Agent.
	// +optional
	FilterByFieldExpectedValue string `json:"filterByFieldExpectedValue,omitempty"`
}

type DomainExtractionGroupingStrategySpec struct {
	// IsGroupingEnabled is the flag to enable Client-side grouping strategy in Integration Agent.
	// +required
	IsGroupingEnabled string `json:"isGroupingEnabled"`

	// GroupByField is the primary group-by field to be used for Client-side grouping strategy in Integration Agent.
	// +optional
	GroupByField string `json:"groupByField,omitempty"`

	// GroupByIdentifier is the field to be used to identify last element in group for Client-side grouping strategy in Integration Agent.
	// +optional
	GroupByIdentifier string `json:"groupByIdentifier,omitempty"`

	// GroupByIdentifierExpectedValue is the field value to be used to identify last element in group for Client-side grouping strategy in Integration Agent.
	// +optional
	GroupByIdentifierExpectedValue string `json:"groupByIdentifierExpectedValue,omitempty"`

	// GroupOperatorType is the matching condition to be used to identify last element in group for Client-side grouping strategy in Integration Agent.
	// +optional
	GroupOperatorType string `json:"groupOperatorType,omitempty"`
}

type DomainExtractionStrategiesSpec struct {
	// +required
	FilterStrategy DomainExtractionFilterStrategySpec `json:"filterStrategy"`

	// +required
	GroupingStrategy DomainExtractionGroupingStrategySpec `json:"groupingStrategy"`
}

type DomainExtractionParametersSpec struct {
	// ConnectionType is the mode to fetch data from S/4Hana source system.
	// +required
	ConnectionType string `json:"connectionType"`

	// GroupEntitySetName is the ODATA Path Entity Set Group from S/4Hana source system.
	// +required
	GroupEntitySetName string `json:"entityGroupName"`

	// EntitySetName is the ODATA Path Entity Set from S/4Hana source system.
	// +required
	EntitySetName string `json:"entitySetName"`

	// Filters is the filters applied to Entity Set from S/4Hana source system.
	// +optional
	Filters string `json:"filters,omitempty"`

	// SelectFields are the select fields applied to Entity Set from S/4Hana source system.
	// +required
	SelectFields string `json:"selectFields"`

	// PaginationField is the pagination field applied to Entity Set from S/4Hana source system.
	// +required
	PaginationField string `json:"paginationField"`

	// ExpandEntities are the list of Expandable entities applied to Entity Set from S/4Hana source system.
	// +optional
	ExpandEntities string `json:"expandEntities,omitempty"`

	// SplitPacketsBy defines how the pagination field applied to Entity Set does chunking from S/4Hana source system.
	// +required
	SplitPacketsBy string `json:"splitPacketsBy"`

	// IsDeltaRelevant defines if the Entity Set is relevant for delta extraction from S/4Hana source system.
	// +required
	IsDeltaRelevant string `json:"isDeltaRelevant"`

	// IsDeltaTimeRangeConsidered defines if the Entity Set considers delta timestamp range relevant for
	// delta extraction from S/4Hana source system.
	// +required
	IsDeltaTimeRangeConsidered string `json:"isDeltaTimeRangeConsidered"`

	// DomainExtractionStrategies defines different extraction strategies for fetching data from S/4Hana source system.
	// +required
	DomainExtractionStrategies DomainExtractionStrategiesSpec `json:"domainExtractionStrategies"`

	// +optional
	AdditionalProperties AdditionalProperties `json:"additionalProperties,omitempty"`
}

type AdditionalProperties struct {
	// +optional
	SapClient string `json:"sapClient,omitempty"`

	// +optional
	SapLanguage string `json:"sapLanguage,omitempty"`
}

type FrameworkParametersSpec struct {
	// PacketGenSchedule is the schedule of the wq-producer replication job.
	// +required
	PacketGenSchedule string `json:"packetGenSchedule"`

	// JobProcessSchedule is the schedule of the wq-processor replication job.
	// +required
	JobProcessSchedule string `json:"jobProcessSchedule"`

	// +required
	DeltaLoadThreshold string `json:"deltaLoadThreshold"`

	// +required
	Parallelism string `json:"parallelism"`

	// +required
	NoOfJobs string `json:"noOfJobs"`

	// +required
	MaxNoOfRecordsInEachPacket string `json:"maxNoOfRecordsInEachPacket"`

	// +required
	IsSourceApiSSLDisabled string `json:"isSourceApiSSLDisabled"`

	// +required
	OAuthURL SecretValueFromSource `json:"oauthUrl"`

	// +required
	OAuthScpUser SecretValueFromSource `json:"oauthScpUser"`

	// +required
	OAuthScpPassword SecretValueFromSource `json:"oauthScpPassword"`

	// +required
	OAuthScpClientID SecretValueFromSource `json:"oauthScpClientID"`

	// +required
	OAuthScpClientSecret SecretValueFromSource `json:"oauthScpClientSecret"`
}

type IntegrationScenarioSpec struct {
	// TenantId is the S/4Hana source system tenant id.
	// +required
	TenantId string `json:"tenantId"`

	// RootObjectType is the root object e.g. Product fetched from S/4Hana source system.
	// +required
	RootObjectType string `json:"rootObjectType"`

	// +required
	DomainExtractionParameters DomainExtractionParametersSpec `json:"domainExtractionParameters"`

	// +required
	FrameworkParameters FrameworkParametersSpec `json:"frameworkParameters"`

	// ServiceAccoutName is the name of the ServiceAccount that will be used to run the Integration
	// Framework Scheduler Deployment.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// Resource limits and Request specifications of the Receive Adapter Deployment
	Resources IntegrationScenarioResourceSpec `json:"resources,omitempty"`
}

type IntegrationScenarioRequestsSpec struct {
	ResourceCPU    string `json:"cpu,omitempty"`
	ResourceMemory string `json:"memory,omitempty"`
}

type IntegrationScenarioLimitsSpec struct {
	ResourceCPU    string `json:"cpu,omitempty"`
	ResourceMemory string `json:"memory,omitempty"`
}

type IntegrationScenarioResourceSpec struct {
	Requests IntegrationScenarioRequestsSpec `json:"requests,omitempty"`
	Limits   IntegrationScenarioLimitsSpec   `json:"limits,omitempty"`
}

type SecretValueFromSource struct {
	// The Secret key to select from.
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

func IntegrationScenarioDeploySource(namespace, integrationScenarioName, tenantId string) string {
	return fmt.Sprintf("/apis/v1/namespaces/%s/integrationScenarios/%s#%s", namespace, integrationScenarioName, tenantId)
}

type IntegrationScenarioStatus struct {
	// inherits duck/v1alpha1 Status, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last processed by the controller.
	// * Conditions - the latest available observations of a resource's current state.
	duckv1.Status `json:",inline"`
}

func (s *IntegrationScenario) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("IntegrationScenario")
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// IntegrationScenarioList contains a list of IntegrationScenarios.
type IntegrationScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items 			[]IntegrationScenario `json:"items"`
}