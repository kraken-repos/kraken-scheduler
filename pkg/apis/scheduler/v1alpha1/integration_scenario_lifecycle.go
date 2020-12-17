package v1alpha1

import (
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"knative.dev/pkg/apis"
)

const (
	IntegrationScenarioConditionDeployed apis.ConditionType = "Deployed"

	IntegrationScenarioConditionResources apis.ConditionType = "ResourcesCorrect"
)

var IntegrationScenarioCondSet = apis.NewLivingConditionSet(
	IntegrationScenarioConditionDeployed,
)

func (s *IntegrationScenario) GetConditionSet() apis.ConditionSet {
	return IntegrationScenarioCondSet
}

func (s *IntegrationScenarioStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return IntegrationScenarioCondSet.Manage(s).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (s *IntegrationScenarioStatus) IsReady() bool {
	return IntegrationScenarioCondSet.Manage(s).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (s *IntegrationScenarioStatus) InitializeConditions() {
	IntegrationScenarioCondSet.Manage(s).InitializeConditions()
}

func CronJobIsAvailable(c *batchv1beta1.CronJobStatus, def bool) bool {
	return true
}

// MarkDeployed sets the condition that the scheduler has been deployed.
func (s *IntegrationScenarioStatus) MarkDeployed(cs []batchv1beta1.CronJob)  {
	for _, c := range cs {
		if CronJobIsAvailable(&c.Status, false) {
			IntegrationScenarioCondSet.Manage(s).MarkTrue(IntegrationScenarioConditionDeployed)
		} else {
			IntegrationScenarioCondSet.Manage(s).MarkFalse(IntegrationScenarioConditionDeployed, "CronJobsUnavailable", "The CronJob '%s' is unavailable.", c.Name)
		}
	}
}

// MarkDeploying sets the condition that the scheduler is deploying.
func (s *IntegrationScenarioStatus) MarkDeploying(reason, messageFormat string, messageA ...interface{})  {
	IntegrationScenarioCondSet.Manage(s).MarkUnknown(IntegrationScenarioConditionDeployed, reason, messageFormat, messageA...)
}

// MarkNotDeployed sets the condition that the scheduler has not been deployed.
func (s *IntegrationScenarioStatus) MarkNotDeployed(reason, messageFormat string, messageA ...interface{}) {
	IntegrationScenarioCondSet.Manage(s).MarkFalse(IntegrationScenarioConditionDeployed, reason, messageFormat, messageA...)
}

func (s *IntegrationScenarioStatus) MarkResourcesCorrect() {
	IntegrationScenarioCondSet.Manage(s).MarkTrue(IntegrationScenarioConditionResources)
}

func (s *IntegrationScenarioStatus) MarkResourcesIncorrect(reason, messageFormat string, messageA ...interface{}) {
	IntegrationScenarioCondSet.Manage(s).MarkFalse(IntegrationScenarioConditionResources, reason, messageFormat, messageA...)
}

