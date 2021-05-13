package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	batchv1beta1Listers "k8s.io/client-go/listers/batch/v1beta1"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	eventingkafkav1beta1 "knative.dev/eventing-kafka/pkg/apis/sources/v1beta1"
	bindingskafkav1beta1 "knative.dev/eventing-kafka/pkg/apis/bindings/v1beta1"
	knEventingKafkaClientSet "knative.dev/eventing-kafka/pkg/client/clientset/versioned"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	pkgreconciler "knative.dev/pkg/reconciler"
	rediscache "kraken.dev/kraken-scheduler/pkg/cache"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingversioned "knative.dev/serving/pkg/client/clientset/versioned"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/cache"
	"kraken.dev/kraken-scheduler/pkg/client/clientset/versioned"
	reconcilerintegrationscenario "kraken.dev/kraken-scheduler/pkg/client/injection/reconciler/scheduler/v1alpha1/integrationscenario"
	listers "kraken.dev/kraken-scheduler/pkg/client/listers/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/messagebroker"
	"kraken.dev/kraken-scheduler/pkg/reconciler/resources"
	"kraken.dev/kraken-scheduler/pkg/slack"
	"strconv"
	"strings"
	"time"
)

const (
	waitQueueProducerImageEnvVar 		  = "WAITQUEUE_PRODUCER_IMAGE"
	waitQueueProcessorImageEnvVar 		  = "WAITQUEUE_PROCESSOR_IMAGE"
	schemaRegistryStoreCredsEnvVar		  = "SCHEMA_REGISTRY_STORE_CREDS"
	integrationSchedulerDeploymentCreated = "IntegrationSchedulerDeploymentCreated"
	integrationSchedulerDeploymentUpdated = "IntegrationSchedulerDeploymentUpdated"
	integrationSchedulerDeploymentFailed  = "IntegrationSchedulerDeploymentFailed"
	component							  = "integration-framework-scheduler"
	componentClusterEnvVar				  = "K8S_CLUSTER"
	componentNamespaceEnvVar			  = "K8S_NAMESPACE"
	kafkaBrokerEnvVar					  = "KAFKA_BROKER_URL"
	kafkaBrokerSASLUserEnvVar			  = "KAFKA_BROKER_SASL_USER"
	kafkaBrokerSASLPasswordEnvVar    	  = "KAFKA_BROKER_SASL_PASSWORD"
	tenantOnboardingTopicEnvVar			  = "TENANT_ONBOARDING_TOPIC"
	intScnOnboardingTopicEnvVar			  = "INT_SCENARIO_ONBOARDING_TOPIC"
	redisUrlEnvVar						  = "CACHE_URL"
)

var (
	deploymentGVK = batchv1beta1.SchemeGroupVersion.WithKind("CronJob")
)

// newDeploymentCreated makes a new reconciler event with event type Normal, and
// reason IntegrationSchedulerDeploymentCreated.
func newDeploymentCreated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, integrationSchedulerDeploymentCreated, "IntegrationScheduler created deployment: \"%s/%s\"", namespace, name)
}

// deploymentUpdated makes a new reconciler event with event type Normal, and
// reason IntegrationSchedulerDeploymentUpdated.
func deploymentUpdated(namespace, name string) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, integrationSchedulerDeploymentUpdated, "IntegrationScheduler updated deployment: \"%s/%s\"", namespace, name)
}

// newDeploymentFailed makes a new reconciler event with event type Warning, and
// reason IntegrationSchedulerDeploymentFailed.
func newDeploymentFailed(namespace, name string, err error) pkgreconciler.Event {
	return pkgreconciler.NewEvent(corev1.EventTypeNormal, integrationSchedulerDeploymentFailed, "IntegrationScheduler failed to create deployment: \"%s/%s\", %w", namespace, name, err)
}

type Reconciler struct {
	// KubeClientSet allows us to talk to the k8s for core APIs
	KubeClientSet 	   kubernetes.Interface
	KnServingClientSet knservingversioned.Interface
	KnEventingKafkaClientSet knEventingKafkaClientSet.Interface

	waitQueueProducerImage   string
	waitQueueProcessorImage  string
	schemaRegistryStoreCreds string


	scenarioLister  listers.IntegrationScenarioLister
	deploymentLister batchv1beta1Listers.CronJobLister

	scenarioClientSet versioned.Interface
	loggingContext     context.Context
	loggingConfig	   *logging.Config
	metricsConfig	   *metrics.ExporterOptions

	SlackClient *slack.Client
	CentralCacheUrl string
	ClusterId   string
	IsPrinted   bool
	IsRedisEmbedded bool
}

// Check that our Reconciler implements Interface
var _ reconcilerintegrationscenario.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.IntegrationScenario) pkgreconciler.Event {
	src.Status.InitializeConditions()

	err := r.provisionKafkaTopic(ctx, src)
	if err != nil {
		logging.FromContext(ctx).Error("error while creating Kafka Topics", zap.Any("integrationScenarioScheduler", err))
	}

	isDemoEnabled, err := strconv.ParseBool(src.Spec.DomainExtractionParameters.IsDemoEnabled)
	redis := r.getRedisExternalIp(ctx, src)

	svcUrl := ""
	if !isDemoEnabled {

		schemaRegistryRequired, err := strconv.ParseBool(src.Spec.DomainExtractionParameters.DomainSchemaRegistryProps.SchemaRegistryRequired)

		if schemaRegistryRequired {
			logging.FromContext(ctx).Info("Register schema for Integration Scenario " + src.Spec.RootObjectType)
			svcUrl, err = r.createSchemaRegistry(ctx, src)
			if err != nil {
				logging.FromContext(ctx).Error("error while creating Schema Registry", zap.Any("integrationScenarioScheduler", err))
			}
		}

		redisCli := rediscache.RedisCli{
			Host: strings.Split(redis, ":")[0],
			Port: strings.Split(redis, ":")[1],
		}

		imagePullPolicy, found := redisCli.GetMapEntry(r.ClusterId + "-" + src.Namespace + "-image", "ImagePullPolicy")
		if !found {
			imagePullPolicy = corev1.PullIfNotPresent
		}

		err = r.createTransformer(ctx, src, imagePullPolicy)
		if err != nil {
			logging.FromContext(ctx).Error("Unable to create the Kraken Transformer", zap.Error(err))
			return err
		}

		err = r.createKafkaSource(ctx, src)
		if err != nil {
			logging.FromContext(ctx).Error("Unable to create the Knative Kafka Source", zap.Error(err))
			return err
		}
	}

	integrationSchedulers, err := r.createIntegrationScenarioDeployers(ctx, svcUrl, redis, r.ClusterId, src)
	if err != nil {
		logging.FromContext(ctx).Error("Unable to create the integration scenario schedulers", zap.Error(err))

		if !r.IsPrinted {
			err = r.SlackClient.SendMessage("Deployed integration scenario: " + src.Name + " in namespace: " + src.Namespace)
			if err != nil {
				logging.FromContext(ctx).Error("Failed to post to slack", zap.Error(err))
			}
		}

		if isDemoEnabled {
			r.IsPrinted = true
		}

		return err
	}
	src.Status.MarkDeployed(integrationSchedulers)

	if !isDemoEnabled {
		err = r.SlackClient.SendMessage("Deployed integration scenario: " + src.Name + " in namespace: " + src.Namespace)
		if err != nil {
			logging.FromContext(ctx).Error("Failed to post to slack", zap.Error(err))
		}
	}

	return nil
}

func (r *Reconciler) getKafkaSecret(ctx context.Context, src *v1alpha1.IntegrationScenario) (*corev1.Secret, error) {
	brokerSecretName := src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef.Name

	secret, err := r.KubeClientSet.CoreV1().Secrets(src.Namespace).Get(ctx, brokerSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (r *Reconciler) provisionKafkaTopic(ctx context.Context, src *v1alpha1.IntegrationScenario) error {
	brokerSecretKey := src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef.Key
	userSecretKey := src.Spec.FrameworkParameters.EventLogUser.SecretKeyRef.Key
	pwdSecretKey := src.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef.Key

	secret, err := r.getKafkaSecret(ctx, src)
	if err != nil {
		return err
	}

	kafkaClient := messagebroker.KafkaClient{
		BootstrapServers: string(secret.Data[brokerSecretKey]),
		SASLUser: string(secret.Data[userSecretKey]),
		SASLPassword: string(secret.Data[pwdSecretKey]),
	}

	err = kafkaClient.Initialize(ctx)
	if err != nil {
		return err
	}

	if kafkaClient.AdminClient != nil {
		defer func() {
			if err = kafkaClient.AdminClient.Close(); err != nil {
				panic(err)
			}
		}()
	}

	parallelism, _ := strconv.Atoi(src.Spec.FrameworkParameters.Parallelism)
	if parallelism%2 == 0 {
		parallelism = parallelism/2
	}

	kafkaClient.CreateTopic(ctx, src.Spec.RootObjectType, int32(parallelism))

	return nil
}

func (r *Reconciler) removeKafkaTopic(ctx context.Context, src *v1alpha1.IntegrationScenario) error {
	brokerSecretName := src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef.Name
	brokerSecretKey := src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef.Key
	userSecretKey := src.Spec.FrameworkParameters.EventLogUser.SecretKeyRef.Key
	pwdSecretKey := src.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef.Key

	secret, err := r.KubeClientSet.CoreV1().Secrets(src.Namespace).Get(ctx, brokerSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	kafkaClient := messagebroker.KafkaClient{
		BootstrapServers: string(secret.Data[brokerSecretKey]),
		SASLUser: string(secret.Data[userSecretKey]),
		SASLPassword: string(secret.Data[pwdSecretKey]),
	}

	err = kafkaClient.Initialize(ctx)
	if err != nil {
		return err
	}

	if kafkaClient.AdminClient != nil {
		defer func() {
			if err = kafkaClient.AdminClient.Close(); err != nil {
				panic(err)
			}
		}()
	}

	kafkaClient.DeleteTopic(ctx, src.Spec.RootObjectType)

	return nil
}

func (r *Reconciler) getRedisExternalIp(ctx context.Context, src *v1alpha1.IntegrationScenario) string {
	if r.IsRedisEmbedded {
		_, err := r.KubeClientSet.CoreV1().Services(src.Namespace).Get(ctx, "redis", metav1.GetOptions{})
		if err != nil {
			return ""
		}
		return "redis." + src.Namespace + ".svc.cluster.local:6379"
	} else {
		redisSecret, err := r.KubeClientSet.CoreV1().Secrets(src.Namespace).Get(ctx, "redis-secret", metav1.GetOptions{})
		if err != nil {
			return ""
		}

		host := string(redisSecret.Data["host"])
		port := string(redisSecret.Data["port"])
		return fmt.Sprintf("%s:%s", host, port)
	}
}

func (r *Reconciler) FinalizeKind(ctx context.Context, src *v1alpha1.IntegrationScenario) pkgreconciler.Event {
	logging.FromContext(ctx).Info("hit the finalizer on deletion" + src.Spec.RootObjectType)

	pods, err := r.KubeClientSet.CoreV1().Pods(src.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, strings.ToLower(src.Spec.RootObjectType) + "-image-builder") ||
			strings.Contains(pod.Name, strings.ToLower(src.Spec.RootObjectType) + "-schema-updater") ||
			strings.Contains(pod.Name, "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType)) {
			name := pod.Name
			var gracePeriod int64 = 0
			var deletePropagation = metav1.DeletePropagationForeground
			err = r.KubeClientSet.CoreV1().Pods(src.Namespace).Delete(ctx, name, metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
				PropagationPolicy: &deletePropagation,
			})
			if err != nil {
				logging.FromContext(ctx).Info("delete pod " + name + "for " + src.Spec.RootObjectType)
				return err
			}
		}
	}

	knservices, err := r.KnServingClientSet.ServingV1().Services(src.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, knservice := range knservices.Items {
		if strings.Contains(knservice.Name, "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType)) {
			name := knservice.Name
			var gracePeriod int64 = 0
			var deletePropagation = metav1.DeletePropagationForeground
			err = r.KnServingClientSet.ServingV1().Services(src.Namespace).Delete(ctx, name, metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
				PropagationPolicy: &deletePropagation,
			})

			if err != nil {
				logging.FromContext(ctx).Info("delete Kn Service " + name + "for " + src.Spec.RootObjectType)
				return err
			}
		}
	}

	cronJobs, err := r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, cronJob := range cronJobs.Items {
		if strings.Contains(cronJob.Name, strings.ToLower(src.Spec.RootObjectType)) {
			name := cronJob.Name
			var gracePeriod int64 = 0
			var deletePropagation = metav1.DeletePropagationForeground
			err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Delete(ctx, name, metav1.DeleteOptions{
				GracePeriodSeconds: &gracePeriod,
				PropagationPolicy: &deletePropagation,
			})
			if err != nil {
				logging.FromContext(ctx).Info("delete cronJob " + name + "for " + src.Spec.RootObjectType)
				return err
			}
		}
	}

	err = r.KubeClientSet.CoreV1().ConfigMaps(src.Namespace).
		Delete(ctx, strings.ToLower(src.Spec.RootObjectType) + "-integration-scenario", metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	err = r.deleteKafkaSource(ctx, src)
	if err != nil {
		return err
	}

	err = r.deleteTransformer(ctx, src)
	if err != nil {
		return err
	}

	err = r.removeKafkaTopic(ctx, src)
	if err != nil {
		return err
	}

	err = r.RemoveRedisQueues(ctx, src)
	if err != nil {
		return err
	}

	err = r.SlackClient.SendMessage("Removed integration scenario: " + src.Name + " in namespace: " + src.Namespace)
	if err != nil {
		logging.FromContext(ctx).Error("Failed to post to slack", zap.Error(err))
	}

	return nil
}

func (r *Reconciler) RemoveRedisQueues(ctx context.Context, src *v1alpha1.IntegrationScenario) error {
	cacheUrl := r.getRedisExternalIp(ctx, src)
	client := cache.Client{
		Host: strings.Split(cacheUrl, ":")[0],
		Port: strings.Split(cacheUrl, ":")[1],
	}

/*	keys := []string{r.ClusterId + "-" + src.Namespace + "-k8swaitqueue-" + src.Spec.RootObjectType,
		r.ClusterId + "-" + src.Namespace + "-k8sworkqueue-" + src.Spec.RootObjectType}*/

	keys := []string{r.ClusterId + "-" + src.Namespace + "-" + "k8swaitqueue-" + src.Spec.RootObjectType,
		r.ClusterId + "-" + src.Namespace + "-" + "k8sworkqueue-" + src.Spec.RootObjectType}

	err := client.DeleteKeysFromCache(ctx, keys)
	if err != nil {
		logging.FromContext(ctx).Info("cannot delete redis objects for " + src.Spec.RootObjectType)
		return err
	}

	delMap := r.ClusterId + "-" + src.Namespace + "-deltaTimestamp"

	res, err := client.GetAllKeysFromMapCache(ctx, delMap)
	if err != nil {
		logging.FromContext(ctx).Info("cannot delete deltaTimestamp for " + src.Spec.RootObjectType)
		return err
	}

	for k, _ := range res {
		entitySetNames := strings.Split(src.Spec.DomainExtractionParameters.EntitySetName, ";")

		for _, entitySetName := range entitySetNames {
			if strings.Contains(k, src.Spec.TenantId) && strings.Contains(k, entitySetName) {
				err = client.DeleteKeyFromMapCache(ctx, delMap, k)
				if err != nil {
					logging.FromContext(ctx).Info("cannot delete deltaTimestamp for " + src.Spec.RootObjectType)
					return err
				}
				break
			}
		}
	}

	redisCli := rediscache.RedisCli{
		Host: strings.Split(r.CentralCacheUrl, ":")[0],
		Port: strings.Split(r.CentralCacheUrl, ":")[1],
	}

	jobTimeTracker := r.ClusterId + "-" + src.Namespace + "-jobTimeTracker" //map[string]time.Time{}
//	jobRestartTracker := r.ClusterId + "-" + src.Namespace + "-jobRestartTracker" //map[string]time.Time{}

	data , err := redisCli.GetAllEntry(jobTimeTracker)
	if err != nil {
		logging.FromContext(ctx).Info("cannot delete informer cache for " + src.Spec.RootObjectType)
		return err
	}
	for key, _ := range data {
		if strings.Contains(key, strings.ToLower(src.Spec.RootObjectType)) {
			err = redisCli.DeleteEntry(jobTimeTracker, key)

			if err != nil {
				logging.FromContext(ctx).Info("cannot delete informer cache for " + src.Spec.RootObjectType)
				return err
			}
		}
	}

/*	data , err = redisCli.GetAllEntry(jobRestartTracker)
	if err != nil {
		logging.FromContext(ctx).Info("cannot delete informer cache for " + src.Spec.RootObjectType)
		return err
	}
	for key, _ := range data {
		if strings.Contains(key, strings.ToLower(src.Spec.RootObjectType)) {
			err = redisCli.DeleteEntry(jobRestartTracker, key)

			if err != nil {
				logging.FromContext(ctx).Info("cannot delete informer cache for " + src.Spec.RootObjectType)
				return err
			}
		}
	}*/

	return nil
}

func checkResourcesStatus(src *v1alpha1.IntegrationScenario) error {
	for _, rsrc := range []struct {
		key   string
		field string
	}{{
		key:   "Request.CPU",
		field: src.Spec.Resources.Requests.ResourceCPU,
	}, {
		key:   "Request.Memory",
		field: src.Spec.Resources.Requests.ResourceMemory,
	}, {
		key:   "Limit.CPU",
		field: src.Spec.Resources.Limits.ResourceCPU,
	}, {
		key:   "Limit.Memory",
		field: src.Spec.Resources.Limits.ResourceMemory,
	}} {
		// In the event the field isn't specified, we assign a default in the integration_scenario
		if rsrc.field != "" {
			if _, err := resource.ParseQuantity(rsrc.field); err != nil {
				src.Status.MarkResourcesIncorrect("Incorrect Resource", "%s: %s, Error: %s", rsrc.key, rsrc.field, err)
				return err
			}
		}
	}
	src.Status.MarkResourcesCorrect()
	return nil
}

func (r *Reconciler) createIntegrationScenarioDeployers(ctx context.Context, svcUrl string, redisIp string, clusterId string, src *v1alpha1.IntegrationScenario) ([]batchv1beta1.CronJob, error) {
	if err := checkResourcesStatus(src); err != nil {
		return nil, err
	}

	loggingConfig, err := logging.ConfigToJSON(r.loggingConfig)
	if err != nil {
		logging.FromContext(ctx).Error("error while converting logging config to JSON", zap.Any("integrationScenarioScheduler", err))
	}

	metricsConfig, err := metrics.OptionsToJSON(r.metricsConfig)
	if err != nil {
		logging.FromContext(ctx).Error("error while converting metrics config to JSON", zap.Any("integrationScenarioScheduler", err))
	}

	integrationScenarioSchedulerArgs := resources.IntegrationScenarioSchedulerArgs{
		WaitQueueProducerImage:  r.waitQueueProducerImage,
		WaitQueueProcessorImage: r.waitQueueProcessorImage,
		Scheduler: 				 src,
		Labels: 				 resources.GetLabels(src.Name, "", ""),
		LoggingConfig: 			 loggingConfig,
		MetricsConfig: 			 metricsConfig,
	}

	logging.FromContext(ctx).Info("Env vars- " + redisIp + "-" + clusterId + "-" + src.Namespace + "-" +
		strconv.FormatBool(r.IsRedisEmbedded))

	logging.FromContext(ctx).Info("Deploying CronJobs for Integration Scenario " + src.Spec.RootObjectType)
	expected := resources.MakeIntegrationScenarioScheduler(&integrationScenarioSchedulerArgs, src.Namespace, svcUrl, redisIp, clusterId)

	jobRunning := false

	waitQueueProducerScheduler, err := r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Get(ctx, expected[0].Name, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		waitQueueProducerScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Create(ctx, &expected[0], metav1.CreateOptions{})
		if err != nil {
			return nil, newDeploymentFailed(waitQueueProducerScheduler.Namespace, waitQueueProducerScheduler.Name, err)
		}
	}

//	time.Sleep(5 * time.Minute)
	waitQueueProcessorScheduler, err := r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Get(ctx, expected[1].Name, metav1.GetOptions{})
	if err == nil {
		jobRunning = true
	}

	if apierrors.IsNotFound(err) {
		waitQueueProcessorScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Create(ctx, &expected[1], metav1.CreateOptions{})
		if err != nil {
			return nil, newDeploymentFailed(waitQueueProcessorScheduler.Namespace, waitQueueProcessorScheduler.Name, err)
		} else {
			jobRunning = true
		}
	}

	if jobRunning {
		return []batchv1beta1.CronJob{*waitQueueProducerScheduler, *waitQueueProcessorScheduler},
			newDeploymentCreated(waitQueueProducerScheduler.Namespace, waitQueueProducerScheduler.Name)
	} else if !jobRunning {
		logging.FromContext(ctx).Error("Unable to get an existing integration scenario scheduler", zap.Error(err))
		return nil, err
	} else if !metav1.IsControlledBy(waitQueueProducerScheduler, src) {
		return nil, fmt.Errorf("deployment %q is not owned by WaitQueueSchedulers %q", waitQueueProducerScheduler.Name, src.Name)
	} else if jobSpecChanged(waitQueueProducerScheduler.Spec.JobTemplate.Spec, expected[0].Spec.JobTemplate.Spec) &&
			jobSpecChanged(waitQueueProcessorScheduler.Spec.JobTemplate.Spec, expected[1].Spec.JobTemplate.Spec) {
		jobRunning = false

		waitQueueProducerScheduler.Spec.JobTemplate.Spec = expected[0].Spec.JobTemplate.Spec
		if waitQueueProducerScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Update(ctx, waitQueueProducerScheduler, metav1.UpdateOptions{}); err != nil {

		}

		waitQueueProcessorScheduler.Spec.JobTemplate.Spec = expected[1].Spec.JobTemplate.Spec
		if waitQueueProcessorScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Update(ctx, waitQueueProcessorScheduler, metav1.UpdateOptions{}); err != nil {

		} else {
			jobRunning = true
		}
		return []batchv1beta1.CronJob{*waitQueueProducerScheduler, *waitQueueProcessorScheduler},
			deploymentUpdated(waitQueueProducerScheduler.Namespace, waitQueueProducerScheduler.Name)
	} else {
		logging.FromContext(ctx).Debug("Reusing existing integration scenario schedulers", zap.Any("waitQueueSchedulers", waitQueueProducerScheduler))
	}

	return []batchv1beta1.CronJob{*waitQueueProducerScheduler, *waitQueueProcessorScheduler}, nil
}

func (r *Reconciler) createSchemaRegistry(ctx context.Context, src *v1alpha1.IntegrationScenario) (string, error) {
	schemaValidatorGenRequired, err := strconv.ParseBool(src.Spec.DomainExtractionParameters.DomainSchemaRegistryProps.SchemaValidatorGenRequired)
	if err != nil {
		schemaValidatorGenRequired = true
	}

	if schemaValidatorGenRequired {
		logging.FromContext(ctx).Info("Get ConfigMap for Integration Scenario " + src.Spec.RootObjectType)
		configMap, err := r.KubeClientSet.CoreV1().ConfigMaps(src.Namespace).Get(ctx, strings.ToLower(src.Spec.RootObjectType) + "-schema", metav1.GetOptions{})

		if err != nil {
			return "", err
		}

		protoFile := configMap.Data[strings.ToLower(src.Spec.RootObjectType) + ".proto"]
		protoValidateFile := configMap.Data[strings.ToLower(src.Spec.RootObjectType) + "-validate.proto"]
		logging.FromContext(ctx).Info(protoFile)

		env := []corev1.EnvVar{
			{
				Name:  "INTEGRATION_SCENARIO",
				Value: src.Spec.RootObjectType,
			},
			{
				Name:  "SCENARIO_SCHEMA",
				Value: protoFile,
			},
			{
				Name:  "SCENARIO_VALIDATION_SCHEMA",
				Value: protoValidateFile,
			},
			{
				Name:  "GIT_EMAIL",
				Value: "sbcd90@gmail.com",
			},
			{
				Name:  "GIT_NAME",
				Value: "Subhobrata Dey",
			},
		}

		logging.FromContext(ctx).Info("Create Schema Updater Pod for Integration Scenario " + src.Spec.RootObjectType)
		schemaUpdaterPod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: strings.ToLower(src.Spec.RootObjectType) + "-schema-updater",
				Labels: resources.GetLabels(src.Name, "", ""),
			},
			Spec: corev1.PodSpec{
				RestartPolicy: "Never",
				ServiceAccountName: src.Spec.ServiceAccountName,
				Containers: []corev1.Container{
					{
						Name:  			 strings.ToLower(src.Spec.RootObjectType) + "-schema-updater",
						Image: 			 "harbor.eurekacloud.io/eureka/schema-updater:latest",
						ImagePullPolicy: "Always",
						Env:  			 env,
					},
				},
				ImagePullSecrets: []corev1.LocalObjectReference{
					{
						Name: "docker-registry-secret",
					},
				},
			},
		}
		_, err = r.KubeClientSet.CoreV1().Pods(src.Namespace).Create(ctx, &schemaUpdaterPod, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
		time.Sleep(1 * time.Minute)

		logging.FromContext(ctx).Info("Create Image Builder Pod for Integration Scenario " + src.Spec.RootObjectType)
		imageBuilderPod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: strings.ToLower(src.Spec.RootObjectType) + "-image-builder",
				Labels: resources.GetLabels(src.Name, "", ""),
			},
			Spec: corev1.PodSpec{
				RestartPolicy:      "Never",
				ServiceAccountName: src.Spec.ServiceAccountName,
				Containers: []corev1.Container{
					{
						Name:            strings.ToLower(src.Spec.RootObjectType) + "-image-builder",
						Image:           "gcr.io/kaniko-project/executor:latest",
						ImagePullPolicy: "Always",
						Args: []string{
							"--dockerfile=./Dockerfile",
							"--context=git://" + r.schemaRegistryStoreCreds + "@github.com/sbcd90/kraken-schema-validator.git#refs/heads/" + src.Spec.RootObjectType,
							"--destination=harbor.eurekacloud.io/eureka/kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "kaniko-secret",
								MountPath: "/kaniko/.docker",
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "kaniko-secret",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "docker-registry-secret",
								Items: []corev1.KeyToPath{
									{
										Key:  ".dockerconfigjson",
										Path: "config.json",
									},
								},
							},
						},
					},
				},
			},
		}

		_, err = r.KubeClientSet.CoreV1().Pods(src.Namespace).Create(ctx, &imageBuilderPod, metav1.CreateOptions{})
		if err != nil {
			return "", err
		}
		time.Sleep(5 * time.Minute)
	}

	schemaValidatorEnv := []corev1.EnvVar{
		{
			Name: "KAFKA_BROKERS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.EventLogUser.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef,
			},
		},
		{
			Name: "SCHEMA_REGISTRY_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.SchemaRegistryEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "SCHEMA_REGISTRY_CREDS",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.SchemaRegistryCreds.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_TOPIC",
			Value: src.Spec.RootObjectType + "ProcessingTopic",
		},
	}

	logging.FromContext(ctx).Info("Create Schema Validator Pod for Integration Scenario " + src.Spec.RootObjectType)
	schemaValidatorPod := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType),
			Labels: resources.GetLabels(src.Name, "", ""),
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: src.Spec.ServiceAccountName,
							Containers: []corev1.Container{
								{
									Name:            "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType),
									Image:           "harbor.eurekacloud.io/eureka/kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType) + ":latest",
									ImagePullPolicy: "Always",
									Env:             schemaValidatorEnv,
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
/*	schemaValidatorPod1 := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType),
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			ServiceAccountName: src.Spec.ServiceAccountName,
			Containers: []corev1.Container{
				{
					Name:  			 "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType),
					Image: 			 "docker.io/sbcd90/kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType) + ":latest",
					ImagePullPolicy: "Always",
					Env:  			 schemaValidatorEnv,
				},
			},
		},
	}*/
	_, err = r.KnServingClientSet.ServingV1().Services(src.Namespace).Create(ctx, &schemaValidatorPod, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	time.Sleep(2 * time.Minute)

	schemaValidatorSvc, err :=
		r.KubeClientSet.CoreV1().Services(src.Namespace).Get(ctx, "kraken-schema-validator-" + strings.ToLower(src.Spec.RootObjectType),
			metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	logging.FromContext(ctx).Info("service ->" + schemaValidatorSvc.Spec.ExternalName, "")
	return schemaValidatorSvc.Spec.ExternalName, nil
}

func (r *Reconciler) deleteTransformer(ctx context.Context, src *v1alpha1.IntegrationScenario) error {
	logging.FromContext(ctx).Info("Delete Kraken Transformer")

	serviceList, err := r.KnServingClientSet.ServingV1().Services(src.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, service := range serviceList.Items {
		if strings.Contains(service.Name, strings.ToLower(src.Spec.RootObjectType) + "-integration-transformer") {
			err = r.KnServingClientSet.ServingV1().Services(src.Namespace).Delete(ctx, service.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) createTransformer(ctx context.Context, src *v1alpha1.IntegrationScenario, imagePullPolicy corev1.PullPolicy) error {
	logging.FromContext(ctx).Info("Create Kraken Transformer")

	partitions, _ := strconv.Atoi(src.Spec.FrameworkParameters.Parallelism)
	if partitions%2 == 0 {
		partitions = partitions/2
	}

	transformerEnv := []corev1.EnvVar{
		{
			Name: "KAFKA_HOST",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.EventLogUser.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_PASSWD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: src.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef,
			},
		},
		{
			Name: "KAFKA_PARTITIONS",
			Value: strconv.Itoa(partitions),
		},
		{
			Name: "FRAMEWORK_PARAMS_REDIS_EMBEDDED",
			Value: strconv.FormatBool(r.IsRedisEmbedded),
		},
		{
			Name: "FRAMEWORK_PARAMS_REDIS_IP",
			Value: r.getRedisExternalIp(ctx, src),
		},
		{
			Name: "FRAMEWORK_PARAMS_NAMESPACE",
			Value: src.Namespace,
		},
		{
			Name: "FRAMEWORK_PARAMS_CLUSTER",
			Value: r.ClusterId,
		},
	}

	transformerPod := servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(src.Spec.RootObjectType) + "-integration-transformer",
			Labels: resources.GetLabels(src.Name, "", ""),
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"autoscaling.knative.dev/class": "kpa.autoscaling.knative.dev",
							"autoscaling.knative.dev/metric": "concurrency",
							"autoscaling.knative.dev/target": "1",
							"autoscaling.knative.dev/minScale": "0",
							"autoscaling.knative.dev/maxScale": src.Spec.FrameworkParameters.Parallelism,
						},
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: src.Spec.ServiceAccountName,
							Containers: []corev1.Container{
								{
									Name:            strings.ToLower(src.Spec.RootObjectType) + "-integration-transformer",
									Image:           "harbor.eurekacloud.io/eureka/kraken-transformer-validator:latest",
									ImagePullPolicy: imagePullPolicy,
									Env:             transformerEnv,
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

	_, err := r.KnServingClientSet.ServingV1().Services(src.Namespace).Create(ctx, &transformerPod, metav1.CreateOptions{})
	return err
}

func (r *Reconciler) deleteKafkaSource(ctx context.Context, src *v1alpha1.IntegrationScenario) error {
	logging.FromContext(ctx).Info("Delete Knative Kafka Source")

	kafkaSourceList, err :=
		r.KnEventingKafkaClientSet.SourcesV1beta1().KafkaSources(src.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, kafkaSource := range kafkaSourceList.Items {
		if strings.Contains(kafkaSource.Name,
			strings.ToLower(src.Spec.RootObjectType) + "-kafka-source") {
			err = r.KnEventingKafkaClientSet.SourcesV1beta1().KafkaSources(src.Namespace).Delete(ctx, kafkaSource.Name, metav1.DeleteOptions{})

			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Reconciler) createKafkaSource(ctx context.Context, src *v1alpha1.IntegrationScenario) error {
	logging.FromContext(ctx).Info("Create Knative Kafka Source")
	brokerSecretName := src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef.Name

	brokerSecretKey := src.Spec.FrameworkParameters.EventLogEndpoint.SecretKeyRef.Key
	userSecretKey := src.Spec.FrameworkParameters.EventLogUser.SecretKeyRef.Key
	pwdSecretKey := src.Spec.FrameworkParameters.EventLogPassword.SecretKeyRef.Key

	secret, err := r.getKafkaSecret(ctx, src)
	if err != nil {
		return err
	}

	parallelism, _ := strconv.Atoi(src.Spec.FrameworkParameters.Parallelism)
	if parallelism%2 == 0 {
		parallelism = parallelism/2
	}

	for instance := 1; instance <= parallelism; instance++ {
		kafkaSourcePod := eventingkafkav1beta1.KafkaSource{
			ObjectMeta: metav1.ObjectMeta{
				Name: strings.ToLower(src.Spec.RootObjectType) + "-kafka-source"+ strconv.Itoa(instance),
				Labels: resources.GetLabels(src.Name, "", ""),
			},
			Spec: eventingkafkav1beta1.KafkaSourceSpec{
				ConsumerGroup: src.Spec.RootObjectType + "ProcessingTopic",
				Topics: []string{src.Spec.RootObjectType + "ProcessingTopic"},
				KafkaAuthSpec: bindingskafkav1beta1.KafkaAuthSpec{
					BootstrapServers: []string{string(secret.Data[brokerSecretKey]), src.Spec.FrameworkParameters.EventBatchThresholdLimit},
					Net: bindingskafkav1beta1.KafkaNetSpec{
						SASL: bindingskafkav1beta1.KafkaSASLSpec{
							Enable: true,
							User: bindingskafkav1beta1.SecretValueFromSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: brokerSecretName,
									},
									Key: userSecretKey,
								},
							},
							Password: bindingskafkav1beta1.SecretValueFromSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: brokerSecretName,
									},
									Key: pwdSecretKey,
								},
							},
						},
						TLS: bindingskafkav1beta1.KafkaTLSSpec{
							Enable: true,
						},
					},
				},
				SourceSpec: duckv1.SourceSpec{
					Sink: duckv1.Destination{
						Ref: &duckv1.KReference{
							APIVersion: "serving.knative.dev/v1",
							Kind: "Service",
							Name: strings.ToLower(src.Spec.RootObjectType) + "-integration-transformer",
						},
					},
				},
			},
		}

		b, _ := json.Marshal(kafkaSourcePod)
		logging.FromContext(ctx).Info("json ->" + string(b))

		_, err := r.KnEventingKafkaClientSet.SourcesV1beta1().KafkaSources(src.Namespace).Create(ctx, &kafkaSourcePod, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func jobSpecChanged(oldJobSpec batchv1.JobSpec, newJobSpec batchv1.JobSpec) bool {
	if !equality.Semantic.DeepDerivative(newJobSpec, oldJobSpec) {
		return true
	}
	if len(oldJobSpec.Template.Spec.Containers) != len(newJobSpec.Template.Spec.Containers) {
		return true
	}
	for i := range newJobSpec.Template.Spec.Containers {
		if !equality.Semantic.DeepEqual(newJobSpec.Template.Spec.Containers[i].Env, oldJobSpec.Template.Spec.Containers[i].Env) {
			return true
		}
	}
	return false
}

func (r *Reconciler) UpdateFromLoggingConfigMap(cfg *corev1.ConfigMap) {
	if cfg != nil {
		delete(cfg.Data, "_example")
	}

	logcfg, err := logging.NewConfigFromConfigMap(cfg)
	if err != nil {
		logging.FromContext(r.loggingContext).Warn("failed to create logging config from configmap", zap.String("cfg.Name", cfg.Name))
		return
	}
	r.loggingConfig = logcfg
	logging.FromContext(r.loggingContext).Info("Update from logging ConfigMap", zap.Any("ConfigMap", cfg))
}

func (r *Reconciler) UpdateFromMetricsConfigMap(cfg *corev1.ConfigMap) {
	if cfg != nil {
		delete(cfg.Data, "_example")
	}

	r.metricsConfig = &metrics.ExporterOptions{
		Domain:    metrics.Domain(),
		Component: component,
		ConfigMap: cfg.Data,
	}
	logging.FromContext(r.loggingContext).Info("Update from metrics ConfigMap", zap.Any("ConfigMap", cfg))
}
