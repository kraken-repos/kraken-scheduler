package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gorilla/mux"
	errors2 "github.com/pkg/errors"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/client/clientset/versioned/scheme"
	"kraken.dev/kraken-scheduler/pkg/messagebroker"
	"kraken.dev/kraken-scheduler/pkg/reconciler/resources"
	"log"
	"net/http"
	"strings"
)

type IntegrationScenarioMsgConsumer struct {
	KubeClientSet 	   kubernetes.Interface
	SchedulerClientSet *SchedulerClientSet
	Consumer 	  	   sarama.Consumer
}

type IntegrationScenarioContent struct {
	Namespace string
	Name 	  string
	Scenario  v1alpha1.IntegrationScenario
	IsCreate  bool
}

func (msgConsumer *IntegrationScenarioMsgConsumer) CreateConsumer(bootstrapServers string,
																 clientId string,
																 clientSecret string,
																 ctx context.Context) error {
	kafkaClient := messagebroker.KafkaClient{
		BootstrapServers: bootstrapServers,
		SASLUser: clientId,
		SASLPassword: clientSecret,
	}
	err := kafkaClient.Initialize(ctx)
	if err != nil {
		return err
	}

	consumer, err := sarama.NewConsumerFromClient(kafkaClient.Client)
	if err != nil {
		return err
	}

	msgConsumer.Consumer = consumer
	return nil
}

func (msgConsumer *IntegrationScenarioMsgConsumer) ListenAndOnboardIntegrationScenario(ctx context.Context, topic string) error {

	if msgConsumer.Consumer != nil {
		partitions, err := msgConsumer.Consumer.Partitions(topic)
		if err != nil {
			return err
		}

		for _, partition := range partitions {
			consumer, err := msgConsumer.Consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
			if err != nil {
				return err
			}

			logging.FromContext(ctx).Info(" Start consuming topic {}, partition {}", topic, partition)
			go func(topic string, consumer sarama.PartitionConsumer) {
				for {
					select {
					case consumerError := <-consumer.Errors():
						err = consumerError.Err
					case msg := <-consumer.Messages():
						logging.FromContext(ctx).Info("Got message: ", string(msg.Value))
						err = msgConsumer.integrationScenarioOnboarding(ctx, msg.Value)
					}
				}
			}(topic, consumer)
		}

		return err
	}

	return errors2.Errorf("Kafka Consumer Object not created")
}

func (msgConsumer *IntegrationScenarioMsgConsumer) integrationScenarioOnboarding(ctx context.Context, data []byte) error {
	logging.FromContext(ctx).Info("Got message: ", string(data))

	integrationScenarioContent := IntegrationScenarioContent{}
	err := json.Unmarshal(data, &integrationScenarioContent)

	if err != nil {
		return err
	}

	if integrationScenarioContent.IsCreate {
		return msgConsumer.createIntegrationScenario(ctx, &integrationScenarioContent)
	} else {
		return msgConsumer.deleteIntegrationScenario(ctx, &integrationScenarioContent)
	}
}

func (msgConsumer *IntegrationScenarioMsgConsumer) createIntegrationScenario(ctx context.Context,
																			integrationScenarioContent *IntegrationScenarioContent) error {
	integrationScenarioBytes, err := json.Marshal(integrationScenarioContent.Scenario)
	logging.FromContext(ctx).Info("Got message integrationScenarioContent: ", integrationScenarioContent.Scenario.Name)

	if err != nil {
		return err
	}

	integrationScenarioCfgMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: integrationScenarioContent.Name,
			Labels: resources.GetLabels(integrationScenarioContent.Name, "", "", ""),
		},
		Data: map[string]string{
			integrationScenarioContent.Name: string(integrationScenarioBytes),
		},
	}
	_, err = msgConsumer.KubeClientSet.CoreV1().ConfigMaps(integrationScenarioContent.Namespace).
		Create(ctx, &integrationScenarioCfgMap, metav1.CreateOptions{})
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}

	err = msgConsumer.SchedulerClientSet.CreateIntegrationScenario(integrationScenarioContent.Namespace, integrationScenarioContent.Scenario)
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}

	return nil
}

func (msgConsumer *IntegrationScenarioMsgConsumer) deleteIntegrationScenario(ctx context.Context,
																			integrationScenarioContent *IntegrationScenarioContent) error {
	err := msgConsumer.SchedulerClientSet.DelIntegrationScenario(integrationScenarioContent.Name, integrationScenarioContent.Namespace)
	if err != nil {
		return err
	}

	err = msgConsumer.KubeClientSet.CoreV1().ConfigMaps(integrationScenarioContent.Namespace).
		Delete(ctx, integrationScenarioContent.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (msgConsumer *IntegrationScenarioMsgConsumer) delIntegrationScenario(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	accessToken := req.Header.Get("Authorization")

	if accessToken == "" {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	if !msgConsumer.SchedulerClientSet.doAuthentication(accessToken) {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	name := vars["name"]

	result := v1alpha1.IntegrationScenario{}

	err := msgConsumer.SchedulerClientSet.clientSet.
		RESTClient().
		Delete().
		Namespace(namespace).
		Resource("integrationscenarios").
		Name(name).
		VersionedParams(&metav1.DeleteOptions{}, scheme.ParameterCodec).
		Do(msgConsumer.SchedulerClientSet.ctx).
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

func (msgConsumer *IntegrationScenarioMsgConsumer) getIntegrationScenario(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	accessToken := req.Header.Get("Authorization")

	if accessToken == "" {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	if !msgConsumer.SchedulerClientSet.doAuthentication(accessToken) {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	name := vars["name"]

	integrationScenario := v1alpha1.IntegrationScenario{}

	err := msgConsumer.SchedulerClientSet.clientSet.
		RESTClient().
		Get().
		Namespace(namespace).
		Resource("integrationscenarios").
		Name(name).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(msgConsumer.SchedulerClientSet.ctx).
		Into(&integrationScenario)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(integrationScenario.Spec)
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

func (msgConsumer *IntegrationScenarioMsgConsumer) listIntegrationScenarios(w http.ResponseWriter, req *http.Request)  {
	vars := mux.Vars(req)

	namespace := vars["namespace"]

	accessToken := req.Header.Get("Authorization")
	if accessToken == "" {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token passed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if !msgConsumer.SchedulerClientSet.doAuthentication(accessToken) {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	opts := metav1.ListOptions{}

	result := v1alpha1.IntegrationScenarioList{}

	err := msgConsumer.SchedulerClientSet.
		clientSet.
		RESTClient().
		Get().
		Namespace(namespace).
		Resource("integrationscenarios").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(msgConsumer.SchedulerClientSet.ctx).
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

func (msgConsumer *IntegrationScenarioMsgConsumer) postIntegrationScenario(w http.ResponseWriter, req *http.Request)  {
	accessToken := req.Header.Get("Authorization")

	if accessToken == "" {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	if !msgConsumer.SchedulerClientSet.doAuthentication(accessToken) {
		logging.FromContext(msgConsumer.SchedulerClientSet.ctx).Info("Access Token auth failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var integrationScenarioContent	IntegrationScenarioContent

	err := json.NewDecoder(req.Body).Decode(&integrationScenarioContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = msgConsumer.createIntegrationScenario(msgConsumer.SchedulerClientSet.ctx, &integrationScenarioContent)
	integrationScenario := integrationScenarioContent.Scenario

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

func(schedulerClientSet *SchedulerClientSet) doAuthentication(accessToken string) bool {
	client := http.Client{}
	req, err := http.NewRequest("GET", "http://oauth-server:9096/oauth/validate", nil)
	if err != nil {
		logging.FromContext(schedulerClientSet.ctx).Info("Connect to auth server failed")
		return false
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", accessToken)

	resp, err := client.Do(req)
	if err != nil {
		logging.FromContext(schedulerClientSet.ctx).Info("Request to auth server failed")
		return false
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.FromContext(schedulerClientSet.ctx).Info("Cannot get response from auth server")
		return false
	}
	bodyStr := string(bodyBytes)

	if strings.Contains(bodyStr, "Valid token") && resp.StatusCode == 200 {
		log.Println("success")
		return true
	}
	return false
}

func StartHTTPServer(msgConsumer *IntegrationScenarioMsgConsumer) {
	r := mux.NewRouter()

	go func() {
		r.HandleFunc("/kraken/postIntegrationScenario/{namespace}", msgConsumer.postIntegrationScenario).
			Methods("POST")
		r.HandleFunc("/kraken/listIntegrationScenarios/{namespace}", msgConsumer.listIntegrationScenarios).
			Methods("GET")
		r.HandleFunc("/kraken/getIntegrationScenario/{namespace}/{name}", msgConsumer.getIntegrationScenario).
			Methods("GET")
		r.HandleFunc("/kraken/deleteIntegrationScenario/{namespace}/{name}", msgConsumer.delIntegrationScenario).
			Methods("DELETE")
		r.HandleFunc("/kraken/listAllTenants", msgConsumer.SchedulerClientSet.listTenants).
			Methods("GET")
		http.ListenAndServe(":443", r)
	}()
}