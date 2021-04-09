package scheduler

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	errors2 "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/messagebroker"
	"kraken.dev/kraken-scheduler/pkg/reconciler/resources"
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
			Labels: resources.GetLabels(integrationScenarioContent.Name, "", ""),
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