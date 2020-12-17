package messagebroker

import (
	"context"
	"github.com/Shopify/sarama"
	"knative.dev/pkg/logging"
	"strings"
	"time"
)

type KafkaClient struct {
	BootstrapServers string
	SASLUser		 string
	SASLPassword	 string
	AdminClient      sarama.ClusterAdmin
}

func (kafkaClient *KafkaClient) Initialize(ctx context.Context) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConfig.Version = sarama.V2_0_1_0
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Net.SASL.Enable = true
	kafkaConfig.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	kafkaConfig.Net.SASL.User = kafkaClient.SASLUser
	kafkaConfig.Net.SASL.Password = kafkaClient.SASLPassword
	kafkaConfig.Admin.Timeout = time.Millisecond * 20000
	kafkaConfig.Net.TLS.Enable = true
	kafkaConfig.Net.TLS.Config = nil

	kafkaAdminClient, err := sarama.NewClusterAdmin(strings.Split(kafkaClient.BootstrapServers, ","), kafkaConfig)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}
	kafkaClient.AdminClient = kafkaAdminClient
}

func (kafkaClient *KafkaClient) CreateTopic(ctx context.Context, rootObjectType string) {
	err := kafkaClient.AdminClient.CreateTopic(rootObjectType + "ProcessingTopic",
		&sarama.TopicDetail{
			NumPartitions: 6,
			ReplicationFactor: 3,
		},
		false,
	)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}

	err = kafkaClient.AdminClient.CreateTopic(rootObjectType + "OutboundTopic",
		&sarama.TopicDetail{
			NumPartitions: 6,
			ReplicationFactor: 3,
		},
		false,
	)
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}
}

func (kafkaClient *KafkaClient) DeleteTopic(ctx context.Context, rootObjectType string) {
	err := kafkaClient.AdminClient.DeleteTopic(rootObjectType + "ProcessingTopic")
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}

/*	err = kafkaClient.AdminClient.DeleteTopic(rootObjectType + "OutboundTopic")
	if err != nil {
		logging.FromContext(ctx).Errorf(err.Error())
	}*/
}