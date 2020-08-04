package resources

const (
	// controllerAgentName is the string used by this controller to identify
	// itself.
	controllerAgentName = "integration-framework-scheduler"
)

func GetLabels(name string) map[string]string {
	return map[string]string{
		"integration.kraken.dev/scheduler":     controllerAgentName,
		"integration.kraken.dev/SchedulerName": name,
	}
}