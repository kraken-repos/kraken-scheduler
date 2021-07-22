package resources

const (
	// controllerAgentName is the string used by this controller to identify
	// itself.
	controllerAgentName = "integration-framework-scheduler"
)

func GetLabels(name string, tenantId string, appTenantId string, rootObjectType string) map[string]string {
	appLabel := appTenantId + "-" + rootObjectType
	if appTenantId == "" {
		appLabel = ""
	}

	return map[string]string{
		"integration.kraken.dev/scheduler":     controllerAgentName,
		"integration.kraken.dev/SchedulerName": name,
		"integration.kraken.dev/appTenantId": appTenantId,
		"app": appLabel,
	}
}