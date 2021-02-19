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

	integrationScenarioInformer := integrationscenarioinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)

	slackClient := slack.Client{
		WebHookURL: "https://hooks.slack.com/services/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",
		Timeout: 5 * time.Second,
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

	schedulerClientSet := StartHTTPServer(ctx, c)

	tenantMsgConsumer := TenantMsgConsumer{
		KubeClientSet: c.KubeClientSet,
		SchedulerClientSet: &schedulerClientSet,
		SlackClient: &slackClient,
	}

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

	return impl
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
