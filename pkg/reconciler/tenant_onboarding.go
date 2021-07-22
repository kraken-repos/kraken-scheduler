package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	errors2 "github.com/pkg/errors"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	rediscache "kraken.dev/kraken-scheduler/pkg/cache"
	podinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/pod"
	"knative.dev/pkg/logging"
	"kraken.dev/kraken-scheduler/pkg/apis/scheduler/v1alpha1"
	"kraken.dev/kraken-scheduler/pkg/messagebroker"
	"kraken.dev/kraken-scheduler/pkg/reconciler/resources"
	"kraken.dev/kraken-scheduler/pkg/slack"
	"strings"
	"time"
)

type TenantMsgConsumer struct {
	KubeClientSet 	   kubernetes.Interface
	Consumer      	   sarama.Consumer
	SchedulerClientSet *SchedulerClientSet
	SlackClient        *slack.Client

	ClusterId string
	CacheUrl  string
	Namespace string
	IsRedisEmbedded bool
}

type Tenant struct {
	TenantId string
	OAuthURL string
	ClientId string
	ClientSecret string

	EventLogEndpoint string
	EventLogUser string
	EventLogPassword string
	SchemaRegistryEndpoint string
	SchemaRegistryCreds string
	ImagePullPolicy string
	TimeoutInMins int64
	IsCreate bool
}

func (tenantMsgConsumer *TenantMsgConsumer) CreateConsumer(bootstrapServers string,
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

	tenantMsgConsumer.Consumer = consumer
	return nil
}

func (tenantMsgConsumer *TenantMsgConsumer) ListenAndOnboardTenant(ctx context.Context, topic string) error {
	if tenantMsgConsumer.IsRedisEmbedded {
		err := tenantMsgConsumer.deployRedis(ctx, &Tenant{TenantId: tenantMsgConsumer.Namespace}, tenantMsgConsumer.Namespace)
		if err != nil {
			return err
		}
	}

	if tenantMsgConsumer.Consumer != nil {
		partitions, err := tenantMsgConsumer.Consumer.Partitions(topic)
		if err != nil {
			return err
		}

		for _, partition := range partitions {
			consumer, err := tenantMsgConsumer.Consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
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
						err = tenantMsgConsumer.tenantOnboarding(ctx, msg.Value)
					}
				}
			}(topic, consumer)
		}

		return err
	}
	return errors2.Errorf("Kafka Consumer Object not created")
}

func (tenantMsgConsumer *TenantMsgConsumer) tenantOnboarding(ctx context.Context, data []byte) error {
	logging.FromContext(ctx).Info("Got message: ", string(data))

	tenant := Tenant{}
	err := json.Unmarshal(data, &tenant)
	logging.FromContext(ctx).Info("Got message: ", tenant.TenantId)

	if err != nil {
		return err
	}

	if tenant.ImagePullPolicy != "" && !tenant.IsCreate {
		logging.FromContext(ctx).Info("Got message for updating ImagePullPolicy: ", tenant.TenantId)
		redisCli := rediscache.RedisCli{
			Host: strings.Split(tenantMsgConsumer.CacheUrl, ":")[0],
			Port: strings.Split(tenantMsgConsumer.CacheUrl, ":")[1],
		}

		err = redisCli.SetMapEntry(tenantMsgConsumer.ClusterId + "-integration-scenarios-" + tenant.TenantId + "-image",
			"ImagePullPolicy", tenant.ImagePullPolicy)
		if err != nil {
			return err
		}
		return nil
	}

	if !tenant.IsCreate {
		logging.FromContext(ctx).Info("Got message for deleting Tenant: ", tenant.TenantId)
		err = tenantMsgConsumer.tenantDeletion(ctx, &tenant)
		if err != nil {
			logging.FromContext(ctx).Info("Error while deleting Tenant: ", err.Error())
			return err
		}
		return nil
	}

	namespace := &corev1.Namespace {
		ObjectMeta: metav1.ObjectMeta{
			Name: "integration-scenarios-" + tenant.TenantId,
		},
	}
	namespace, err = tenantMsgConsumer.KubeClientSet.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}
	logging.FromContext(ctx).Info("Namespace created for Tenant ", tenant.TenantId)

	err = tenantMsgConsumer.deploySecrets(ctx, &tenant, "integration-scenarios-" + tenant.TenantId)
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}

	err = tenantMsgConsumer.deployRedis(ctx, &tenant, "integration-scenarios-" + tenant.TenantId)
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}

	err = tenantMsgConsumer.deployPostgres(ctx, &tenant, "integration-scenarios-" + tenant.TenantId)
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}

	if tenant.TimeoutInMins < 10.0 {
		tenant.TimeoutInMins = 10.0
	}

	err = tenantMsgConsumer.startTenantInformer(ctx, "integration-scenarios-" + tenant.TenantId, tenant.TimeoutInMins)
	if err != nil {
		logging.FromContext(ctx).Error(err.Error())
		return err
	}

	return nil
}

func (tenantMsgConsumer *TenantMsgConsumer) tenantDeletion(ctx context.Context, tenant *Tenant) error {
	namespace := "integration-scenarios-" + tenant.TenantId

	integrationScenarioList, err := tenantMsgConsumer.SchedulerClientSet.ListIntegrationScenarios(namespace)

	for _, is := range integrationScenarioList {
		err = tenantMsgConsumer.SchedulerClientSet.DelIntegrationScenario(is, namespace)
	}
	logging.FromContext(ctx).Info("deleted IntegrationScenarios for Tenant ", tenant.TenantId)

	redisCli := rediscache.RedisCli{
		Host: strings.Split(tenantMsgConsumer.CacheUrl, ":")[0],
		Port: strings.Split(tenantMsgConsumer.CacheUrl, ":")[1],
	}
	mapNames := []string{
		tenantMsgConsumer.ClusterId + "-" + namespace + "-jobTimeTracker",
		tenantMsgConsumer.ClusterId + "-" + namespace + "-jobRestartTracker",
	}
	err = redisCli.DeleteMap(mapNames)
	logging.FromContext(ctx).Info("deleted RedisQueues for Tenant ", tenant.TenantId)

	err = tenantMsgConsumer.KubeClientSet.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("deleted Tenant ", tenant.TenantId)
	return nil
}

func (tenantMsgConsumer *TenantMsgConsumer) deploySecrets(ctx context.Context, tenant *Tenant, namespace string) error {
	krakenSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kraken-secret",
		},
		Data: map[string][]byte{
			"url": []byte(tenant.OAuthURL),
			"clientId": []byte(tenant.ClientId),
			"clientSecret": []byte(tenant.ClientSecret),
			"username": []byte(tenant.ClientId),
			"password": []byte(tenant.ClientSecret),
		},
		Type: corev1.SecretTypeOpaque,
	}
	krakenSecret, err := tenantMsgConsumer.KubeClientSet.CoreV1().Secrets("integration-scenarios-" + tenant.TenantId).Create(ctx, krakenSecret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("kraken-secret created for Tenant ", tenant.TenantId)

	kafkaSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kafka-secret",
		},
		Data: map[string][]byte{
			"eventLogEndpoint": []byte(tenant.EventLogEndpoint),
			"eventLogUser": []byte(tenant.EventLogUser),
			"eventLogPassword": []byte(tenant.EventLogPassword),
			"schemaRegistryEndpoint": []byte(tenant.SchemaRegistryEndpoint),
			"schemaRegistryCreds": []byte(tenant.SchemaRegistryCreds),
		},
		Type: corev1.SecretTypeOpaque,
	}
	kafkaSecret, err = tenantMsgConsumer.KubeClientSet.CoreV1().Secrets("integration-scenarios-" + tenant.TenantId).Create(ctx, kafkaSecret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("kafka-secret created for Tenant ", tenant.TenantId)

	dockerRegistrySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "docker-registry-secret",
		},
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: []byte("{\"auths\":{\"harbor.eurekacloud.io\":{\"username\":\"subhobrata.dey@sap.com\",\"password\":\"bt74qMgu3eGCjJHQ\",\"email\":\"subhobrata.dey@sap.com\",\"auth\":\"c3ViaG9icmF0YS5kZXlAc2FwLmNvbTpidDc0cU1ndTNlR0NqSkhR\"}}}"),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}
	dockerRegistrySecret, err = tenantMsgConsumer.KubeClientSet.CoreV1().Secrets("integration-scenarios-" + tenant.TenantId).Create(ctx, dockerRegistrySecret, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("docker-registry-secret created for Tenant ", tenant.TenantId)

	return nil
}

func (tenantMsgConsumer *TenantMsgConsumer) deployRedis(ctx context.Context, tenant *Tenant, namespace string) error {
	if !tenantMsgConsumer.IsRedisEmbedded {
		redisSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis-secret",
			},
			Data: map[string][]byte{
				"host": []byte(strings.Split(tenantMsgConsumer.CacheUrl, ":")[0]),
				"port": []byte(strings.Split(tenantMsgConsumer.CacheUrl, ":")[1]),
			},
			Type: corev1.SecretTypeOpaque,
		}

		redisSecret, err := tenantMsgConsumer.KubeClientSet.CoreV1().Secrets("integration-scenarios-" + tenant.TenantId).Create(ctx, redisSecret, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		logging.FromContext(ctx).Info("redis-secret created for Tenant ", tenant.TenantId)
	} else {
		replicas := int32(1)

		redisDeployment := &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis",
			},
			Spec: v1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "redis",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "redis",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container {
							{
								Name: "redis-helm-chart",
								Image: "redis:latest",
								ImagePullPolicy: "IfNotPresent",
								Ports: []corev1.ContainerPort{
									{
										Name: "http",
										ContainerPort: 6379,
									},
								},
							},
						},
					},
				},
			},
		}

		redisSvc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "redis",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeNodePort,
				Ports: [] corev1.ServicePort{
					{
						Port: 6379,
						TargetPort: intstr.IntOrString{
							Type: intstr.Int,
							IntVal: 6379,
							StrVal: "6379",
						},
						Protocol: corev1.ProtocolTCP,
						Name: "redis",
					},
				},
				Selector: map[string]string{
					"app": "redis",
				},
			},
		}

		redisDeployment, err := tenantMsgConsumer.KubeClientSet.AppsV1().Deployments(namespace).Create(ctx, redisDeployment, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		logging.FromContext(ctx).Info("redis deployment created for Tenant ", tenant.TenantId)

		redisSvc, err = tenantMsgConsumer.KubeClientSet.CoreV1().Services(namespace).Create(ctx, redisSvc, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		logging.FromContext(ctx).Info("redis service created for Tenant ", tenant.TenantId)
	}

	return nil
}

func (tenantMsgConsumer *TenantMsgConsumer) deployPostgres(ctx context.Context, tenant *Tenant, namespace string) error {
	replicas := int32(1)

	postgresDeployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgresql",
		},
		Spec: v1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "postgresql",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "postgresql",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container {
						{
							Name: "postgresql-helm-chart",
							Image: "postgres:latest",
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								{
									Name: "http",
									ContainerPort: 5432,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "POSTGRES_PASSWORD",
									Value: "postgres",
								},
							},
						},
					},
				},
			},
		},
	}

	postgresSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "postgresql",
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: [] corev1.ServicePort{
				{
					Port: 5432,
					TargetPort: intstr.IntOrString{
						Type: intstr.Int,
						IntVal: 5432,
						StrVal: "5432",
					},
					Protocol: corev1.ProtocolTCP,
					Name: "postgresql",
				},
			},
			Selector: map[string]string{
				"app": "postgresql",
			},
		},
	}

	postgresDeployment, err := tenantMsgConsumer.KubeClientSet.AppsV1().Deployments(namespace).Create(ctx, postgresDeployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("postgres deployment created for Tenant ", tenant.TenantId)

	postgresSvc, err = tenantMsgConsumer.KubeClientSet.CoreV1().Services(namespace).Create(ctx, postgresSvc, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	logging.FromContext(ctx).Info("postgres service created for Tenant ", tenant.TenantId)

	return nil
}

func (tenantMsgConsumer *TenantMsgConsumer) startTenantInformer(ctx context.Context, namespace string, timeoutInMins int64) error {
	redisCli := rediscache.RedisCli{
		Host: strings.Split(tenantMsgConsumer.CacheUrl, ":")[0],
		Port: strings.Split(tenantMsgConsumer.CacheUrl, ":")[1],
	}
/*	redisCli.SetEntry("test1", "hello6", time.Now())
	data, _ := redisCli.GetAllEntry("test")

	logging.FromContext(ctx).Info("Cache print1:")
	for k, v := range data {
		logging.FromContext(ctx).Info(k + "-" + v.String())
	}
	time.Sleep(5 * time.Second)

	redisCli.SetEntry("test1", "hello7", time.Now())
	data, _ = redisCli.GetAllEntry("test1")

	logging.FromContext(ctx).Info("Cache print2:")
	for k, v := range data {
		logging.FromContext(ctx).Info(k + "-" + v.String())
	}
	time.Sleep(5 * time.Second)

	redisCli.DeleteEntry("test1", "hello4")
	data, _ = redisCli.GetAllEntry("test1")

	logging.FromContext(ctx).Info("Cache print4:")
	for k, v := range data {
		logging.FromContext(ctx).Info(k + "-" + v.String())
	}

	data1, _ := redisCli.GetEntry("test1", "hello5")
	logging.FromContext(ctx).Info("Cache print5: " + data1.String())*/

	timeout := time.Duration(timeoutInMins) * time.Minute
	jobTimeTracker := tenantMsgConsumer.ClusterId + "-" + namespace + "-jobTimeTracker" //map[string]time.Time{}
	jobRestartTracker := tenantMsgConsumer.ClusterId + "-" + namespace + "-jobRestartTracker" //map[string]time.Time{}

	podInformer := podinformer.Get(ctx)
	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)

			if pod.GetNamespace() == namespace {
				if (strings.Contains(pod.GetName(), "-prod-") ||
					strings.Contains(pod.GetName(), "-proc-")) &&
					pod.Status.Phase != corev1.PodSucceeded {
					jobKeySplit := strings.Split(pod.GetName(), "-")
					jobKey := jobKeySplit[0] + "-" + namespace

					if _, ok := redisCli.GetEntry(jobTimeTracker, jobKey); !ok {
						err := redisCli.SetEntry(jobTimeTracker, jobKey, pod.GetCreationTimestamp().Time)
						if err != nil {
							logging.FromContext(ctx).Info("Failed to save pod to redis:" + pod.GetName())
						}
					}
				}

				if (strings.Contains(pod.GetName(), "-prod-") ||
					strings.Contains(pod.GetName(), "-proc-")) &&
					pod.Status.Phase == corev1.PodSucceeded {
					jobKeySplit := strings.Split(pod.GetName(), "-")
					jobKey := jobKeySplit[0] + "-" + namespace

					logging.FromContext(ctx).Info("Test pod informer success:" + pod.GetName())
					err := redisCli.DeleteEntry(jobTimeTracker, jobKey)
					if err != nil {
						logging.FromContext(ctx).Info("Failed to delete pod to redis:" + pod.GetName())
					}
				}

				logging.FromContext(ctx).Info("Test pod informer:" + pod.GetNamespace() + "-" +
					pod.GetName() + "-" + string(pod.Status.Phase))
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
//			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)

			if newPod.GetNamespace() == namespace {

				if (strings.Contains(newPod.GetName(), "-prod-") ||
					strings.Contains(newPod.GetName(), "-proc-")) &&
					newPod.Status.Phase != corev1.PodSucceeded {
					jobKeySplit := strings.Split(newPod.GetName(), "-")
					jobKey := jobKeySplit[0] + "-" + namespace

					if _, ok := redisCli.GetEntry(jobTimeTracker, jobKey); !ok {
						err := redisCli.SetEntry(jobTimeTracker, jobKey, newPod.GetCreationTimestamp().Time)
						if err != nil {
							logging.FromContext(ctx).Info("Failed to save pod to redis:" + newPod.GetName())
						}
					}
				}

				if (strings.Contains(newPod.GetName(), "-prod-") ||
					strings.Contains(newPod.GetName(), "-proc-")) &&
					newPod.Status.Phase == corev1.PodSucceeded {
					jobKeySplit := strings.Split(newPod.GetName(), "-")
					jobKey := jobKeySplit[0] + "-" + namespace

					logging.FromContext(ctx).Info("Test pod informer success:" + newPod.GetName())
					err := redisCli.DeleteEntry(jobTimeTracker, jobKey)
					if err != nil {
						logging.FromContext(ctx).Info("Failed to delete pod to redis:" + newPod.GetName())
					}
				}

				logging.FromContext(ctx).Info("Test pod informer:" + newPod.GetNamespace() + "-" +
					newPod.GetName() + "-" + string(newPod.Status.Phase))
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)

			if pod.GetNamespace() == namespace {

				if (strings.Contains(pod.GetName(), "-prod-") ||
					strings.Contains(pod.GetName(), "-proc-")) &&
					pod.Status.Phase != corev1.PodSucceeded {
					jobKeySplit := strings.Split(pod.GetName(), "-")
					jobKey := jobKeySplit[0] + "-" + namespace

					if _, ok := redisCli.GetEntry(jobTimeTracker, jobKey); !ok {
						err := redisCli.SetEntry(jobTimeTracker, jobKey, pod.GetCreationTimestamp().Time)
						if err != nil {
							logging.FromContext(ctx).Info("Failed to save pod to redis:" + pod.GetName())
						}
					}
				}

				if (strings.Contains(pod.GetName(), "-prod-") ||
					strings.Contains(pod.GetName(), "-proc-")) &&
					pod.Status.Phase == corev1.PodSucceeded {
					jobKeySplit := strings.Split(pod.GetName(), "-")
					jobKey := jobKeySplit[0] + "-" + namespace

					logging.FromContext(ctx).Info("Test pod informer success:" + pod.GetName())
					err := redisCli.DeleteEntry(jobTimeTracker, jobKey)
					if err != nil {
						logging.FromContext(ctx).Info("Failed to delete pod to redis:" + pod.GetName())
					}
				}

				logging.FromContext(ctx).Info("Test pod informer:" + pod.GetNamespace() + "-" +
					pod.GetName() + "-" + string(pod.Status.Phase))
			}
		},
	})

	go func() {
		for true {
			jobTimeTrackerMap, err := redisCli.GetAllEntry(jobTimeTracker)
			if err != nil {
				logging.FromContext(ctx).Info("Failed to get pods from redis: " + err.Error())
				continue
			}

			for pod, timestamp := range jobTimeTrackerMap {
				age := time.Since(timestamp).Round(time.Minute)
				if age >= timeout {
					logging.FromContext(ctx).Info("Test pod informer error:" + pod)
					integrationScenarioParts := strings.Split(pod, "-")

					integrationScenarioName := integrationScenarioParts[0] + "-integration-scenario"

					cfgMap, err := tenantMsgConsumer.KubeClientSet.CoreV1().ConfigMaps(namespace).
						Get(ctx, integrationScenarioName, metav1.GetOptions{})
					if err != nil {
						logging.FromContext(ctx).Info("Failed to load Configmap: " + integrationScenarioName +
							"in namespace: " + namespace)
					}
					if cfgMap != nil {
						integrationScenarioJson := cfgMap.Data[integrationScenarioName]
						logging.FromContext(ctx).Info("Loaded Configmap: " + integrationScenarioJson +
							"in namespace: " + namespace)

						err = tenantMsgConsumer.SchedulerClientSet.DelIntegrationScenario(integrationScenarioName, namespace)
						if err != nil {
							logging.FromContext(ctx).Info("Failed to delete Integration Scenario: " + integrationScenarioName +
								"in namespace: " + namespace)
						} else {
							err = tenantMsgConsumer.SlackClient.SendMessage("Integration failed for object: " +
								integrationScenarioParts[0] + " for tenant: " + namespace)
							if err != nil {
								logging.FromContext(ctx).Error("Failed to post to slack", zap.Error(err))
							}

							err = tenantMsgConsumer.SlackClient.SendMessage("Deleted Integration Scenario: " + integrationScenarioName +
								"in namespace: " + namespace + ". Replication will auto-start" +
								" in " + fmt.Sprintf("%f", timeout.Minutes()) + " mins. Please debug the issue meantime from logs.")
							if err != nil {
								logging.FromContext(ctx).Error("Failed to post to slack", zap.Error(err))
							}
						}

						err = redisCli.DeleteEntry(jobTimeTracker, pod)
						if err != nil {
							logging.FromContext(ctx).Info("Failed to delete pod to redis:" + pod)
						}
						logging.FromContext(ctx).Info("Deleted Integration Scenario: " + integrationScenarioName +
							"in namespace: " + namespace + ". Replication will auto-start" +
							" in " + fmt.Sprintf("%f", timeout.Minutes()) + " mins. Please debug the issue meantime from logs.")

						integrationScenarioCfgMap := corev1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name: integrationScenarioName,
								Labels: resources.GetLabels(integrationScenarioName, "", "", ""),
							},
							Data: map[string]string{
								integrationScenarioName: integrationScenarioJson,
							},
						}

						_, err = tenantMsgConsumer.KubeClientSet.CoreV1().ConfigMaps(namespace).
							Create(ctx, &integrationScenarioCfgMap, metav1.CreateOptions{})
						if err != nil {
							logging.FromContext(ctx).Error(err.Error())
						}

						err = redisCli.SetEntry(jobRestartTracker, integrationScenarioName+":"+namespace, time.Now())
						if err != nil {
							logging.FromContext(ctx).Info("Failed to save pod to redis:" + integrationScenarioName + ":" + namespace)
						}
					}
				}
			}
			time.Sleep(time.Duration(timeoutInMins/2) * time.Minute)
		}
	}()

	go func() {
		for true {
			jobRestartTrackerMap, err := redisCli.GetAllEntry(jobRestartTracker)
			if err != nil {
				logging.FromContext(ctx).Info("Failed to get pods from redis: " + err.Error())
				continue
			}

			for pod, timestamp := range jobRestartTrackerMap {
				age := time.Since(timestamp).Round(time.Minute)
				if age >= timeout {
					logging.FromContext(ctx).Info("Test pod restart: ", pod)
					integrationScenarioParts := strings.Split(pod, ":")

					integrationScenario := integrationScenarioParts[0]
					integrationScenarioNs := integrationScenarioParts[1]

					if integrationScenarioNs == namespace {
						cfgMap, err := tenantMsgConsumer.KubeClientSet.CoreV1().ConfigMaps(namespace).
							Get(ctx, integrationScenario, metav1.GetOptions{})

						if err != nil {
							logging.FromContext(ctx).Info("Failed to load Configmap: " + integrationScenario +
								"in namespace: " + integrationScenarioNs)
						}

						if cfgMap != nil {
							integrationScenarioJson := cfgMap.Data[integrationScenario]
							logging.FromContext(ctx).Info("Loaded Configmap: " + integrationScenarioJson +
								"in namespace: " + integrationScenarioNs)

							integrationScenarioObj := v1alpha1.IntegrationScenario{}

							err := json.Unmarshal([]byte(integrationScenarioJson), &integrationScenarioObj)
							if err != nil {
								logging.FromContext(ctx).Info(err.Error())
							}

							err = tenantMsgConsumer.SchedulerClientSet.CreateIntegrationScenario(integrationScenarioNs, integrationScenarioObj)
							if err != nil {
								logging.FromContext(ctx).Info("Failed to create Integration Scenario: " + integrationScenario +
									"in namespace: " + integrationScenarioNs)
							} else {
								err = tenantMsgConsumer.SlackClient.SendMessage("Re-Created Integration Scenario for object: " +
									integrationScenario + " for tenant: " + namespace)
								if err != nil {
									logging.FromContext(ctx).Error("Failed to post to slack", zap.Error(err))
								}
							}

							err = redisCli.DeleteEntry(jobRestartTracker, pod)
							if err != nil {
								logging.FromContext(ctx).Info("Failed to delete pod to redis:" + pod)
							}
							logging.FromContext(ctx).Info("Re-Created Integration Scenario: " + integrationScenario +
								"in namespace: " + namespace)
						} else {
							logging.FromContext(ctx).Info("Failed to load Configmap: " + integrationScenario +
								"in namespace: " + integrationScenarioNs)

							err = redisCli.DeleteEntry(jobRestartTracker, pod)
							if err != nil {
								logging.FromContext(ctx).Info("Failed to delete pod to redis:" + pod)
							}
						}
					}
				}
			}
			time.Sleep(time.Duration(timeoutInMins/2) * time.Minute)
		}
	}()

	return nil
}