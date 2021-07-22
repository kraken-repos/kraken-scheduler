package resources

import (
	"encoding/json"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"knative.dev/pkg/kmeta"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	rediscache "kraken.dev/kraken-scheduler/pkg/cache"
	"strconv"
	"strings"
)

type IntegrationScenarioSchedulerArgs struct {
	WaitQueueProducerImage  string
	WaitQueueProcessorImage string
	Scheduler				*v1alpha1.IntegrationScenario
	Labels					map[string]string
	MetricsConfig           string
	LoggingConfig			string
}

func GenerateFixedName(owner metav1.Object, prefix string) string {
	uid := string(owner.GetUID())

	if !strings.HasPrefix(uid, "-") {
		uid = "-" + uid
	}

	prefix = strings.TrimSuffix(prefix, "-")
	pl := validation.DNS1123LabelMaxLength - len(uid)
	if pl > len(prefix) {
		pl = len(prefix)
	}
	return prefix[:pl] + uid
}

func SerializeArray(data []v1alpha1.DomainExtractionParametersSpec) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func MakeIntegrationScenarioScheduler(args *IntegrationScenarioSchedulerArgs, currNs string, svcUrl string, redisIp string, clusterId string, timezone string) []batchv1beta1.CronJob {
	partitions, _ := strconv.Atoi(args.Scheduler.Spec.FrameworkParameters.Parallelism)
	if partitions%2 == 0 {
		partitions = partitions/2
	}

	env := []corev1.EnvVar{
		{
			Name:  "S4HANA_TENANT_ID",
			Value: args.Scheduler.Spec.TenantId,
		},
		{
			Name:  "S4HANA_ROOT_OBJECT_TYPE",
			Value: args.Scheduler.Spec.RootObjectType,
		},
		{
			Name:  "S4HANA_APP_TENANT_ID",
			Value: args.Scheduler.Spec.AppTenantId,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_CONNECTION_TYPE",
			Value: args.Scheduler.Spec.DomainExtractionParameters.ConnectionType,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_ENTITY_GROUP_NAME",
			Value: args.Scheduler.Spec.DomainExtractionParameters.GroupEntitySetName,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_ENTITY_SET_NAME",
			Value: args.Scheduler.Spec.DomainExtractionParameters.EntitySetName,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_FILTERS",
			Value: args.Scheduler.Spec.DomainExtractionParameters.Filters,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_SELECT_FIELDS",
			Value: args.Scheduler.Spec.DomainExtractionParameters.SelectFields,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_PAGINATION_FIELD",
			Value: args.Scheduler.Spec.DomainExtractionParameters.PaginationField,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_EXPAND_ENTITIES",
			Value: args.Scheduler.Spec.DomainExtractionParameters.ExpandEntities,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_ENTITY_SETS_DELTA_FLAG",
			Value: args.Scheduler.Spec.DomainExtractionParameters.IsDeltaRelevant,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_SPLIT_PACKETS_BY",
			Value: args.Scheduler.Spec.DomainExtractionParameters.SplitPacketsBy,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_ENTITY_SETS_TIME_RANGE_DELTA_FLAG",
			Value: args.Scheduler.Spec.DomainExtractionParameters.IsDeltaTimeRangeConsidered,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_FILTER_BY_FIELD",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.FilterStrategy.FilterByField,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_FILTER_BY_KEY_EXPECTED_VALUE",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.FilterStrategy.FilterByFieldExpectedValue,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_FILTER_BY_ENABLED",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.FilterStrategy.IsFilterEnabled,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_GROUP_BY_FIELD",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.GroupingStrategy.GroupByField,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_GROUP_BY_IDENTIFIER",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.GroupingStrategy.GroupByIdentifier,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_GROUP_BY_IDENTIFIER_EXPECTED_VALUE",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.GroupingStrategy.GroupByIdentifierExpectedValue,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_GROUP_BY_ENABLED",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.GroupingStrategy.IsGroupingEnabled,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_GROUP_OPERATOR_TYPE",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.GroupingStrategy.GroupOperatorType,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_DEMO_ENABLED",
			Value: args.Scheduler.Spec.DomainExtractionParameters.IsDemoEnabled,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_USE_JOB_TIMESTAMP",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainExtractionStrategies.UseJobTimestampStrategy.UseJobTimestamp,
		},
		{
			Name:  "DOMAIN_EXTRACTOR_START_PACKETS_FROM",
			Value: args.Scheduler.Spec.DomainExtractionParameters.DomainStartPacketsFrom,
		},
		{
			Name:  "FRAMEWORK_PARAMS_PARALLELISM",
			Value: strconv.Itoa(partitions),
		},
		{
			Name:  "FRAMEWORK_PARAMS_NO_OF_PACKETS",
			Value: args.Scheduler.Spec.FrameworkParameters.NoOfJobs,
		},
		{
			Name:  "FRAMEWORK_PARAMS_NO_OF_RECORDS",
			Value: args.Scheduler.Spec.FrameworkParameters.MaxNoOfRecordsInEachPacket,
		},
		{
			Name:  "FRAMEWORK_PARAMS_SOURCE_API_SSL_DISABLED",
			Value: args.Scheduler.Spec.FrameworkParameters.IsSourceApiSSLDisabled,
		},
		{
			Name:  "FRAMEWORK_PARAMS_DELTA_LOAD_THRESHOLD",
			Value: args.Scheduler.Spec.FrameworkParameters.DeltaLoadThreshold,
		},
		{
			Name: "FRAMEWORK_PARAMS_EVENT_LOG_ENDPOINT",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "FRAMEWORK_PARAMS_EVENT_LOG_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.EventLogUser.SecretKeyRef,
			},
		},
		{
			Name: "FRAMEWORK_PARAMS_EVENT_LOG_PASSWD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef,
			},
		},
		{
			Name: "FRAMEWORK_PARAMS_SCHEMA_VALIDATOR_URL",
			Value: svcUrl,
		},
		{
			Name: "FRAMEWORK_PARAMS_REDIS_IP",
			Value: redisIp,
		},
		{
			Name: "FRAMEWORK_PARAMS_TIMEZONE",
			Value: timezone,
		},
		{
			Name: "FRAMEWORK_PARAMS_INTERVAL_CALLS",
			Value: args.Scheduler.Spec.FrameworkParameters.IntervalBetweenCalls,
		},
		{
			Name:  "ADDL_PARAMS_SAP_CLIENT",
			Value: args.Scheduler.Spec.DomainExtractionParameters.AdditionalProperties.SapClient,
		},
		{
			Name:  "ADDL_PARAMS_SAP_LANGUAGE",
			Value: args.Scheduler.Spec.DomainExtractionParameters.AdditionalProperties.SapLanguage,
		},
		{
			Name:  "NAME",
			Value: args.Scheduler.Name,
		},
		{
			Name:  "NAMESPACE",
			Value: args.Scheduler.Namespace,
		},
		{
			Name:  "K_LOGGING_CONFIG",
			Value: args.LoggingConfig,
		},
		{
			Name:  "K_METRICS_CONFIG",
			Value: args.MetricsConfig,
		},
		{
			Name: "FRAMEWORK_PARAMS_NAMESPACE",
			Value: currNs,
		},
		{
			Name: "FRAMEWORK_PARAMS_CLUSTER",
			Value: clusterId,
		},
	}

	isBasicAuthUsed, _ := strconv.ParseBool(args.Scheduler.Spec.FrameworkParameters.IsBasicAuthUsed)

	if isBasicAuthUsed {
		env = append(env,
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_BASIC_AUTH_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.BasicAuthUser.SecretKeyRef,
			},
		},
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_BASIC_AUTH_PASSWD",
			ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.BasicAuthPassword.SecretKeyRef,
		},
		})
	} else {
		env = append(env,
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_OAUTH_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.OAuthURL.SecretKeyRef,
			},
		},
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_OAUTH_SCP_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.OAuthScpUser.SecretKeyRef,
			},
		},
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_OAUTH_SCP_PASSWD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.OAuthScpPassword.SecretKeyRef,
			},
		},
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_OAUTH_CLIENTID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.OAuthScpClientID.SecretKeyRef,
			},
		},
		corev1.EnvVar {
		Name: "FRAMEWORK_PARAMS_OAUTH_CLIENTSECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.OAuthScpClientSecret.SecretKeyRef,
			},
		})
	}

	/*_ := []corev1.EnvVar{
		{
			Name: "KAFKA_BROKERS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.EventLogUser.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef,
			},
		},
		{
			Name: "SCHEMA_REGISTRY_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.SchemaRegistryEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "SCHEMA_REGISTRY_CREDS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: args.Scheduler.Spec.FrameworkParameters.SchemaRegistryEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_TOPIC",
			Value: args.Scheduler.Spec.RootObjectType + "SchemaProcessingTopic",
		},
	}*/

	RequestResourceCPU, err := resource.ParseQuantity(args.Scheduler.Spec.Resources.Requests.ResourceCPU)
	if err != nil {
		RequestResourceCPU = resource.MustParse("250m")
	}
	RequestResourceMemory, err := resource.ParseQuantity(args.Scheduler.Spec.Resources.Requests.ResourceMemory)
	if err != nil {
		RequestResourceMemory = resource.MustParse("512Mi")
	}
	LimitResourceCPU, err := resource.ParseQuantity(args.Scheduler.Spec.Resources.Limits.ResourceCPU)
	if err != nil {
		LimitResourceCPU = resource.MustParse("250m")
	}
	LimitResourceMemory, err := resource.ParseQuantity(args.Scheduler.Spec.Resources.Limits.ResourceMemory)
	if err != nil {
		LimitResourceMemory = resource.MustParse("512Mi")
	}

	res := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    RequestResourceCPU,
			corev1.ResourceMemory: RequestResourceMemory,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    LimitResourceCPU,
			corev1.ResourceMemory: LimitResourceMemory,
		},
	}

	parallelism, err := strconv.Atoi(args.Scheduler.Spec.FrameworkParameters.Parallelism)
	var parallel int32 = 1
	if err != nil {
		parallel = 1
	}
	parallel = int32(parallelism)

	rootObjectType := strings.ToLower(args.Scheduler.Spec.RootObjectType)
	if len(rootObjectType) >= 10 {
		rootObjectType = rootObjectType[:10]
	}

	redisCli := rediscache.RedisCli{
		Host: strings.Split(redisIp, ":")[0],
		Port: strings.Split(redisIp, ":")[1],
	}

	imagePullPolicy, found := redisCli.GetMapEntry(clusterId + "-" + currNs + "-image", "ImagePullPolicy")
	if !found {
		imagePullPolicy = corev1.PullIfNotPresent
	}

	cronJobForWaitQueueProducer := batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: GenerateFixedName(args.Scheduler, fmt.Sprintf(strings.ToLower(rootObjectType) + "-prod")),
			Namespace: args.Scheduler.Namespace,
			Labels: args.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Scheduler),
			},
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: args.Scheduler.Spec.FrameworkParameters.PacketGenSchedule,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: args.Labels,
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "Never",
							ServiceAccountName: args.Scheduler.Spec.ServiceAccountName,
							Containers: []corev1.Container{
								{
									Name:  	   		 strings.ToLower(args.Scheduler.Spec.RootObjectType) + "-producer",
									Image: 	   		 args.WaitQueueProducerImage,
									ImagePullPolicy: imagePullPolicy,
									Env:   	   		 env,
									Resources: 		 res,
								},
							},
							ImagePullSecrets: []corev1.LocalObjectReference{
								{
									Name: "docker-registry-secret",
								},
							},
						},
					},
				},
			},
		},
	}

	cronJobForWaitQueueProcessor := batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: GenerateFixedName(args.Scheduler, fmt.Sprintf(strings.ToLower(rootObjectType) + "-proc")),
			Namespace: args.Scheduler.Namespace,
			Labels: args.Labels,
			OwnerReferences: []metav1.OwnerReference{
				*kmeta.NewControllerRef(args.Scheduler),
			},
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule: args.Scheduler.Spec.FrameworkParameters.JobProcessSchedule,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Parallelism: &parallel,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: args.Labels,
						},
						Spec: corev1.PodSpec{
							RestartPolicy: "Never",
							ServiceAccountName: args.Scheduler.Spec.ServiceAccountName,
							Containers: []corev1.Container{
								{
									Name:  	   		 strings.ToLower(args.Scheduler.Spec.RootObjectType) + "-processor",
									Image: 	   		 args.WaitQueueProcessorImage,
									ImagePullPolicy: imagePullPolicy,
									Env:   	   		 env,
									Resources: 		 res,
								},
								/*{
									Name:  	   		 "kraken-schema-validator-" + strings.ToLower(args.Scheduler.Spec.RootObjectType),
									Image: 	   		 "docker.io/sbcd90/kraken-schema-validator-" + strings.ToLower(args.Scheduler.Spec.RootObjectType) + ":latest",
									ImagePullPolicy: "Always",
									Env:   	   		 sideCarEnv,
									Resources: 		 res,
								},*/
							},
							ImagePullSecrets: []corev1.LocalObjectReference{
								{
									Name: "docker-registry-secret",
								},
								/*{
									Name: "regcred",
								},*/
							},
						},
					},
				},
			},
		},
	}

	return []batchv1beta1.CronJob{cronJobForWaitQueueProducer, cronJobForWaitQueueProcessor}
}
