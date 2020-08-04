package scheduler

import (
	"context"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	integrationscenarioclient "kraken.dev/kraken-scheduler/pkg/client/injection/client"
	integrationscenarioinformer "kraken.dev/kraken-scheduler/pkg/client/injection/informers/scheduler/v1alpha1/integrationscenario"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/batch/v1beta1/cronjob"
	"os"

	"knative.dev/pkg/controller"
	"kraken.dev/kraken-scheduler/pkg/client/injection/reconciler/scheduler/v1alpha1/integrationscenario"
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

	integrationScenarioInformer := integrationscenarioinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)

	c := &Reconciler{
		KubeClientSet: 	   		 kubeclient.Get(ctx),
		scenarioClientSet: 		 integrationscenarioclient.Get(ctx),
		scenarioLister:    		 integrationScenarioInformer.Lister(),
		deploymentLister:  		 deploymentInformer.Lister(),
		waitQueueProducerImage:  wQProducerImage,
		waitQueueProcessorImage: wQProcessorImage,
		loggingContext:			 ctx,
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
	return impl
}
