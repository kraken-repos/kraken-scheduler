package scheduler

import (
	"context"
	errors2 "github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"
	knEventingKafkaClientSet "knative.dev/eventing-kafka/pkg/client/injection/client"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/batch/v1beta1/cronjob"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	knservingclient "knative.dev/serving/pkg/client/injection/client"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	integrationscenarioclient "kraken.dev/kraken-scheduler/pkg/client/injection/client"
	integrationscenarioinformer "kraken.dev/kraken-scheduler/pkg/client/injection/informers/scheduler/v1alpha1/integrationscenario"
	"kraken.dev/kraken-scheduler/pkg/client/injection/reconciler/scheduler/v1alpha1/integrationscenario"
	"kraken.dev/kraken-scheduler/pkg/slack"
	"os"
	"time"
)

func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	wQProducerImage, defined := os.LookupEnv(waitQueueProducerImageEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", waitQueueProducerImageEnvVar)
		return nil
	}

	wQProcessorImage, defined := os.LookupEnv(waitQueueProcessorImageEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", waitQueueProcessorImageEnvVar)
		return nil
	}

	schemaRegistryStoreCreds, defined := os.LookupEnv(schemaRegistryStoreCredsEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", schemaRegistryStoreCredsEnvVar)
		return nil
	}

	k8sCluster, defined := os.LookupEnv(componentClusterEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", componentClusterEnvVar)
		return nil
	}

	k8sNamespace, defined := os.LookupEnv(componentNamespaceEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", componentNamespaceEnvVar)
		return nil
	}

	k8sTimezone, defined := os.LookupEnv(componentTimezoneEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", componentTimezoneEnvVar)
		return nil
	}

	integrationScenarioInformer := integrationscenarioinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)

	slackClient := slack.Client{
		WebHookURL: "https://hooks.slack.com/services/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",
		Timeout: 5 * time.Second,
	}

	isRedisEmbedded := false
	cacheUrl, err := GetRedisCredentialsFromEnvVar(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf("required cache credentials not found, moving to embedded mode", err.Error())
		isRedisEmbedded = true

		cacheUrl = "redis." + k8sNamespace + ".svc.cluster.local:6379"
	}

	c := &Reconciler{
		KubeClientSet: 	   		  kubeclient.Get(ctx),
		KnServingClientSet:  	  knservingclient.Get(ctx),
		KnEventingKafkaClientSet: knEventingKafkaClientSet.Get(ctx),
		scenarioClientSet: 		  integrationscenarioclient.Get(ctx),
		scenarioLister:    		  integrationScenarioInformer.Lister(),
		deploymentLister:  		  deploymentInformer.Lister(),
		waitQueueProducerImage:   wQProducerImage,
		waitQueueProcessorImage:  wQProcessorImage,
		schemaRegistryStoreCreds: schemaRegistryStoreCreds,
		loggingContext:			  ctx,
		SlackClient: 			  &slackClient,
		ClusterId:     			  k8sCluster,
		Timezone:  				  k8sTimezone,
		CentralCacheUrl: 		  cacheUrl,
		IsRedisEmbedded:  		  isRedisEmbedded,
		IsPrinted: 				  false,
	}

	impl := integrationscenario.NewImpl(ctx, c)

	logging.FromContext(ctx).Info("Setting up integration-framework-scheduler event handlers")

	integrationScenarioInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("IntegrationScenario")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	cmw.Watch(logging.ConfigMapName(), c.UpdateFromLoggingConfigMap)
	cmw.Watch(metrics.ConfigMapName(), c.UpdateFromMetricsConfigMap)

	clientSet, err := BootstrapServer(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}

	schedulerClientSet := SchedulerClientSet{
		ctx: 		ctx,
		clientSet:  clientSet,
	}

	tenantMsgConsumer := TenantMsgConsumer{
		KubeClientSet: c.KubeClientSet,
		SchedulerClientSet: &schedulerClientSet,
		SlackClient: &slackClient,
		ClusterId: k8sCluster,
		CacheUrl: cacheUrl,
		Namespace: k8sNamespace,
		IsRedisEmbedded: isRedisEmbedded,
	}

	integrationScenarioMsgConsumer := IntegrationScenarioMsgConsumer{
		KubeClientSet: c.KubeClientSet,
		SchedulerClientSet: &schedulerClientSet,
	}
	StartHTTPServer(&integrationScenarioMsgConsumer)

	brokerUrl, brokerUser, brokerPassword, err := GetKafkaCredentialsFromEnvVar(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf("required kafka credentials not found", err.Error())
		return nil
	}

	err = tenantMsgConsumer.CreateConsumer(brokerUrl, brokerUser, brokerPassword, ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil
	}

	err = integrationScenarioMsgConsumer.CreateConsumer(brokerUrl, brokerUser, brokerPassword, ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil
	}

	tenantOnboardingTopic, defined := os.LookupEnv(tenantOnboardingTopicEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", tenantOnboardingTopicEnvVar)
		return nil
	}

	err = tenantMsgConsumer.ListenAndOnboardTenant(ctx, tenantOnboardingTopic)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil
	}

	integrationScenarioTopic, defined := os.LookupEnv(intScnOnboardingTopicEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", intScnOnboardingTopicEnvVar)
		return nil
	}

	err = integrationScenarioMsgConsumer.ListenAndOnboardIntegrationScenario(ctx, integrationScenarioTopic)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil
	}

	return impl
}

func GetRedisCredentialsFromEnvVar(ctx context.Context) (string, error) {
	redisUrl, defined := os.LookupEnv(redisUrlEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", redisUrlEnvVar)
		return "", errors2.Errorf("required environment variable '%s' not defined", redisUrlEnvVar)
	}

	return redisUrl, nil
}

func GetKafkaCredentialsFromEnvVar(ctx context.Context) (string, string, string, error) {
	brokerUrl, defined := os.LookupEnv(kafkaBrokerEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", kafkaBrokerEnvVar)
		return "", "", "", errors2.Errorf("required environment variable '%s' not defined", kafkaBrokerEnvVar)
	}

	brokerUser, defined := os.LookupEnv(kafkaBrokerSASLUserEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", kafkaBrokerSASLUserEnvVar)
		return "", "", "", errors2.Errorf("required environment variable '%s' not defined", kafkaBrokerSASLUserEnvVar)
	}

	brokerPassword, defined := os.LookupEnv(kafkaBrokerSASLPasswordEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable '%s' not defined", kafkaBrokerSASLPasswordEnvVar)
		return "", "", "", errors2.Errorf("required environment variable '%s' not defined", kafkaBrokerSASLPasswordEnvVar)
	}

	return brokerUrl, brokerUser, brokerPassword, nil
}
