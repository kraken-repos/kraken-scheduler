package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/logging"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/client/clientset/versioned/scheme"
	schedulerv1alpha1 "kraken.dev/kraken-scheduler/pkg/client/clientset/versioned/typed/scheduler/v1alpha1"
	"net/http"
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

func (schedulerClientSet *SchedulerClientSet) deleteIntegrationScenario(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	name := vars["name"]

	result := v1alpha1.IntegrationScenario{}

	err := schedulerClientSet.clientSet.
		RESTClient().
		Delete().
		Namespace(namespace).
		Resource("integrationscenarios").
		Name(name).
		VersionedParams(&metav1.DeleteOptions{}, scheme.ParameterCodec).
		Do(schedulerClientSet.ctx).
		Into(&result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)

	_, err = fmt.Fprintf(w, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (schedulerClientSet *SchedulerClientSet) getIntegrationScenario(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	name := vars["name"]

	integrationScenario := v1alpha1.IntegrationScenario{}

	err := schedulerClientSet.clientSet.
		RESTClient().
		Get().
		Namespace(namespace).
		Resource("integrationscenarios").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(schedulerClientSet.ctx).
		Into(&integrationScenario)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(IntegrationScenarioResp{
		TenantId:       integrationScenario.Spec.TenantId,
		AppTenantId:    integrationScenario.Spec.RootObjectType,
		RootObjectType: integrationScenario.Spec.RootObjectType,
		Name: 			integrationScenario.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)

	_, err = fmt.Fprintf(w, string(resp))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (schedulerClientSet *SchedulerClientSet) listIntegrationScenarios(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	opts := metav1.ListOptions{}

	result := v1alpha1.IntegrationScenarioList{}

	err := schedulerClientSet.clientSet.
		RESTClient().
		Get().
		Namespace(namespace).
		Resource("integrationscenarios").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(schedulerClientSet.ctx).
		Into(&result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var integrationScenarioListResp []IntegrationScenarioResp
	for _, is := range result.Items {
		integrationScenarioListResp =
			append(integrationScenarioListResp, IntegrationScenarioResp{
				TenantId: is.Spec.TenantId,
				AppTenantId: is.Spec.AppTenantId,
				RootObjectType: is.Spec.RootObjectType,
				Name: is.Name,
			})
	}

	resp, err := json.Marshal(integrationScenarioListResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	_, err = fmt.Fprintf(w, string(resp))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (schedulerClientSet *SchedulerClientSet) postIntegrationScenario(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	var integrationScenario	v1alpha1.IntegrationScenario

	err := json.NewDecoder(req.Body).Decode(&integrationScenario)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := v1alpha1.IntegrationScenario{}

	err = schedulerClientSet.clientSet.
		RESTClient().
		Post().
		Namespace(namespace).
		Resource("integrationscenarios").
		Body(&integrationScenario).
		Do(schedulerClientSet.ctx).
		Into(&result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(IntegrationScenarioResp{
		TenantId:       integrationScenario.Spec.TenantId,
		AppTenantId:    integrationScenario.Spec.RootObjectType,
		RootObjectType: integrationScenario.Spec.RootObjectType,
		Name: integrationScenario.Name,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)

	_, err = fmt.Fprintf(w, string(resp))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func bootstrapServer(ctx context.Context) (*schedulerv1alpha1.SchedulerV1alpha1Client, error) {
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

func StartHTTPServer(ctx context.Context, reconciler *Reconciler) {
	r := mux.NewRouter()

	clientSet, err := bootstrapServer(ctx)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}

	schedulerClientSet := &SchedulerClientSet{
		ctx: 		ctx,
		clientSet:  clientSet,
		reconciler: reconciler,
	}

	go func() {
		r.HandleFunc("/postIntegrationScenario/{namespace}", schedulerClientSet.postIntegrationScenario).
			Methods("POST")
		r.HandleFunc("/listIntegrationScenarios/{namespace}", schedulerClientSet.listIntegrationScenarios).
			Methods("GET")
		r.HandleFunc("/getIntegrationScenario/{namespace}/{name}", schedulerClientSet.getIntegrationScenario).
			Methods("GET")
		r.HandleFunc("/deleteIntegrationScenario/{namespace}/{name}", schedulerClientSet.deleteIntegrationScenario).
			Methods("DELETE")
		http.ListenAndServe(":443", r)
	}()
}