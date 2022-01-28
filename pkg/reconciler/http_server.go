package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/logging"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/client/clientset/versioned/scheme"
	schedulerv1alpha1 "kraken.dev/kraken-scheduler/pkg/client/clientset/versioned/typed/scheduler/v1alpha1"
	"net/http"
	"strings"
)

type SchedulerClientSet struct {
	ctx 	   context.Context
	reconciler *Reconciler
	clientSet  *schedulerv1alpha1.SchedulerV1alpha1Client
}

type IntegrationScenarioResp struct {
	TenantId       string
	AppTenantId    string
	RootObjectType string
	Name		   string
}

func (schedulerClientSet *SchedulerClientSet) listTenants(w http.ResponseWriter, req *http.Request)  {
	accessToken := req.Header.Get("Authorization")

	if accessToken == "" {
		logging.FromContext(schedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	if !schedulerClientSet.doAuthentication(accessToken) {
		logging.FromContext(schedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	nsList, err := schedulerClientSet.reconciler.
		KubeClientSet.
		CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var namespaces []string
	for _, ns := range nsList.Items {
		if strings.HasPrefix(ns.Name, "integration-scenarios") {
			namespaces = append(namespaces, ns.Name)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	resp, err := json.Marshal(namespaces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = fmt.Fprintf(w, string(resp))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (schedulerClientSet *SchedulerClientSet) DelIntegrationScenario(name, namespace string) error {
	result := v1alpha1.IntegrationScenario{}

	return schedulerClientSet.clientSet.
		RESTClient().
		Delete().
		Namespace(namespace).
		Resource("integrationscenarios").
		Name(name).
		VersionedParams(&metav1.DeleteOptions{}, scheme.ParameterCodec).
		Do(schedulerClientSet.ctx).
		Into(&result)
}

func (schedulerClientSet *SchedulerClientSet) ListIntegrationScenarios(namespace string) ([]string, error) {
	result := v1alpha1.IntegrationScenarioList{}

	err := schedulerClientSet.clientSet.
		RESTClient().
		Get().
		Namespace(namespace).
		Resource("integrationscenarios").
		VersionedParams(&metav1.ListOptions{}, scheme.ParameterCodec).
		Do(schedulerClientSet.ctx).
		Into(&result)

	if err != nil {
		return nil, err
	}

	var integrationScenarioList []string
	for _, is := range result.Items {
		integrationScenarioList = append(integrationScenarioList, is.GetName())
	}

	return integrationScenarioList, nil
}

func (schedulerClientSet *SchedulerClientSet) CreateIntegrationScenario(namespace string,
																		integrationScenario v1alpha1.IntegrationScenario) error {
	result := v1alpha1.IntegrationScenario{}

	return schedulerClientSet.clientSet.
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("integrationscenarios").
		Body(&integrationScenario).
		Do(schedulerClientSet.ctx).
		Into(&result)
}

func BootstrapServer(ctx context.Context) (*schedulerv1alpha1.SchedulerV1alpha1Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil, err
	}

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil, err
	}

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &v1alpha1.SchemeGroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	clientSet, err := schedulerv1alpha1.NewForConfig(&crdConfig)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
		return nil, err
	}
	return clientSet, nil
}