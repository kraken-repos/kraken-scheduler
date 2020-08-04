package scheduler

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	batchv1beta1Listers "k8s.io/client-go/listers/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/metrics"
	pkgreconciler "knative.dev/pkg/reconciler"
	"kraken.dev/kraken-scheduler/pkg/reconciler/resources"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	reconcilerintegrationscenario "kraken.dev/kraken-scheduler/pkg/client/injection/reconciler/scheduler/v1alpha1/integrationscenario"
	"kraken.dev/kraken-scheduler/pkg/client/clientset/versioned"
	listers "kraken.dev/kraken-scheduler/pkg/client/listers/scheduler/v1alpha1"
)

const (
	waitQueueProducerImageEnvVar 		  = "WAITQUEUE_PRODUCER_IMAGE"
	waitQueueProcessorImageEnvVar 		  = "WAITQUEUE_PROCESSOR_IMAGE"
	integrationSchedulerDeploymentCreated = "IntegrationSchedulerDeploymentCreated"
	integrationSchedulerDeploymentUpdated = "IntegrationSchedulerDeploymentUpdated"
	integrationSchedulerDeploymentFailed  = "IntegrationSchedulerDeploymentFailed"
	component							  = "integration-framework-scheduler"
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
	KubeClientSet kubernetes.Interface

	waitQueueProducerImage  string
	waitQueueProcessorImage string

	scenarioLister  listers.IntegrationScenarioLister
	deploymentLister batchv1beta1Listers.CronJobLister

	scenarioClientSet versioned.Interface
	loggingContext     context.Context
	loggingConfig	   *logging.Config
	metricsConfig	   *metrics.ExporterOptions
}

// Check that our Reconciler implements Interface
var _ reconcilerintegrationscenario.Interface = (*Reconciler)(nil)

func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.IntegrationScenario) pkgreconciler.Event {
	src.Status.InitializeConditions()

	integrationSchedulers, err := r.createIntegrationScenarioDeployers(ctx, src)
	if err != nil {
		logging.FromContext(ctx).Error("Unable to create the integration scenario schedulers", zap.Error(err))
		return err
	}
	src.Status.MarkDeployed(integrationSchedulers)

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

func (r *Reconciler) createIntegrationScenarioDeployers(ctx context.Context, src *v1alpha1.IntegrationScenario) ([]batchv1beta1.CronJob, error) {
	if err := checkResourcesStatus(src); err != nil {
		return nil, err
	}

	loggingConfig, err := logging.LoggingConfigToJson(r.loggingConfig)
	if err != nil {
		logging.FromContext(ctx).Error("error while converting logging config to JSON", zap.Any("integrationScenarioScheduler", err))
	}

	metricsConfig, err := metrics.MetricsOptionsToJson(r.metricsConfig)
	if err != nil {
		logging.FromContext(ctx).Error("error while converting metrics config to JSON", zap.Any("integrationScenarioScheduler", err))
	}

	integrationScenarioSchedulerArgs := resources.IntegrationScenarioSchedulerArgs{
		WaitQueueProducerImage:  r.waitQueueProducerImage,
		WaitQueueProcessorImage: r.waitQueueProcessorImage,
		Scheduler: 				 src,
		Labels: 				 resources.GetLabels(src.Name),
		LoggingConfig: 			 loggingConfig,
		MetricsConfig: 			 metricsConfig,
	}

	expected := resources.MakeIntegrationScenarioScheduler(&integrationScenarioSchedulerArgs)

	jobRunning := false

	waitQueueProducerScheduler, err := r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Get(expected[0].Name, metav1.GetOptions{})

	if apierrors.IsNotFound(err) {
		waitQueueProducerScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Create(&expected[0])
		if err != nil {
			return nil, newDeploymentFailed(waitQueueProducerScheduler.Namespace, waitQueueProducerScheduler.Name, err)
		}
	}

	waitQueueProcessorScheduler, err := r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Get(expected[1].Name, metav1.GetOptions{})
	if err == nil {
		jobRunning = true
	}

	if apierrors.IsNotFound(err) {
		waitQueueProcessorScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Create(&expected[1])
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
		if waitQueueProducerScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Update(waitQueueProducerScheduler); err != nil {

		}

		waitQueueProcessorScheduler.Spec.JobTemplate.Spec = expected[1].Spec.JobTemplate.Spec
		if waitQueueProcessorScheduler, err = r.KubeClientSet.BatchV1beta1().CronJobs(src.Namespace).Update(waitQueueProcessorScheduler); err != nil {

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
