/*
Copyright 2022 TriggerMesh Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kafkatarget

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	pkgadapter "knative.dev/eventing/pkg/adapter/v2"
	"knative.dev/pkg/logging"

	"github.com/Shopify/sarama"

	"github.com/triggermesh/triggermesh/pkg/apis/targets"
	"github.com/triggermesh/triggermesh/pkg/metrics"
)

// NewTarget adapter implementation
func NewTarget(ctx context.Context, envAcc pkgadapter.EnvConfigAccessor, ceClient cloudevents.Client) pkgadapter.Adapter {
	logger := logging.FromContext(ctx)

	mt := &pkgadapter.MetricTag{
		ResourceGroup: targets.KafkaTargetResource.String(),
		Namespace:     envAcc.GetNamespace(),
		Name:          envAcc.GetName(),
	}

	metrics.MustRegisterEventProcessingStatsView()

	env := envAcc.(*envAccessor)

	var sc sarama.Client
	var err error

	config := sarama.NewConfig()
	tlsCfg := &tls.Config{}

	if env.SALSEnable {
		config.Net.SASL.Enable = env.SALSEnable
		config.Net.SASL.Mechanism = sarama.SASLMechanism(env.SecurityMechanisms)
		config.Net.SASL.User = env.Username
		config.Net.SASL.Password = env.Password
	}

	if env.TLSEnable {
		config.Net.TLS.Enable = env.TLSEnable
		tlsCfg, err = newTLSCertificatesConfig(tlsCfg, env.SSLClientCert, env.SSLClientKey)
		if err != nil {
			logger.Panicw("Could not create the TLS Certificates Config", err)
		}
		tlsCfg = newTLSRootCAConfig(tlsCfg, env.SSLCA)
		config.Net.TLS.Config = tlsCfg
	}

	if env.SecurityMechanisms == "GSSAPI" {
		kerberosConfig := sarama.GSSAPIConfig{
			AuthType:           sarama.KRB5_KEYTAB_AUTH,
			KeyTabPath:         env.KerberosKeytabPath,
			KerberosConfigPath: env.KerberosConfigPath,
			ServiceName:        env.KerberosServiceName,
			Username:           env.KerberosUsername,
			Password:           env.KerberosPassword,
			Realm:              env.KerberosRealm,
			DisablePAFXFAST:    true,
		}
		config.Net.SASL.GSSAPI = kerberosConfig
	}

	config.Producer.Return.Successes = true
	err = config.Validate()
	if err != nil {
		logger.Panicw("Config not valid", err)
	}

	sc, err = sarama.NewClient(
		env.BootstrapServers,
		config,
	)
	if err != nil {
		logger.Panicw("Error creating Sarama Client", err)
	}

	kc, err := sarama.NewSyncProducerFromClient(sc)
	if err != nil {
		logger.Panicw("Error creating Kafka Producer", err)
	}

	return &kafkaAdapter{
		saramaClient:              sc,
		kafkaClient:               kc,
		topic:                     env.Topic,
		createTopicIfMissing:      env.CreateTopicIfMissing,
		flushTimeout:              env.FlushOnExitTimeoutMillisecs,
		topicTimeout:              env.CreateTopicTimeoutMillisecs,
		newTopicPartitions:        env.NewTopicPartitions,
		newTopicReplicationFactor: env.NewTopicReplicationFactor,

		discardCEContext: env.DiscardCEContext,

		ceClient: ceClient,
		logger:   logger,

		sr: metrics.MustNewEventProcessingStatsReporter(mt),
	}
}

var _ pkgadapter.Adapter = (*kafkaAdapter)(nil)

type kafkaAdapter struct {
	saramaClient sarama.Client
	kafkaClient  sarama.SyncProducer
	topic        string

	createTopicIfMissing bool

	flushTimeout              int
	topicTimeout              int
	newTopicPartitions        int
	newTopicReplicationFactor int

	discardCEContext bool

	ceClient cloudevents.Client
	logger   *zap.SugaredLogger

	sr *metrics.EventProcessingStatsReporter
}

func (a *kafkaAdapter) Start(ctx context.Context) error {
	a.logger.Info("Starting Kafka adapter")

	defer func() {
		a.kafkaClient.Close()
	}()

	return a.ceClient.StartReceiver(ctx, a.dispatch)
}

func (a *kafkaAdapter) dispatch(event cloudevents.Event) cloudevents.Result {
	ceTypeTag := metrics.TagEventType(event.Type())
	ceSrcTag := metrics.TagEventSource(event.Source())

	var msgVal []byte

	start := time.Now()
	defer func() {
		a.sr.ReportProcessingLatency(time.Since(start), ceTypeTag, ceSrcTag)
	}()

	if a.discardCEContext {
		msgVal = event.Data()
	} else {
		jsonEvent, err := json.Marshal(event)
		if err != nil {
			a.logger.Errorw("Error marshalling CloudEvent", zap.Error(err))
			a.sr.ReportProcessingError(true, ceTypeTag, ceSrcTag)
			return err
		}
		msgVal = jsonEvent
	}

	msg := &sarama.ProducerMessage{
		Topic: a.topic,
		Key:   sarama.StringEncoder(event.ID()),
		Value: sarama.ByteEncoder(msgVal),
	}

	_, _, err := a.kafkaClient.SendMessage(msg)
	if err != nil {
		a.logger.Errorw("Error producing Kafka message", zap.String("msg", string(msgVal)), zap.Error(err))
		a.sr.ReportProcessingError(true, ceTypeTag, ceSrcTag)
		return err
	}

	return cloudevents.ResultACK
}