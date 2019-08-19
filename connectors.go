package main

import (
	"crypto/tls"
	"fmt"
	"github.com/Shopify/sarama"
	"os"
	"strings"
)

type Clients interface {
	SendMessageSync(body []byte) error
	Close()
}

type Connectors struct {
	producer sarama.SyncProducer
	Name     string
}

func NewClientConnectors(cd ConnectionData) Clients {
	logger.Trace("Creating Kafka message producer")

	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	cfg.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	cfg.Producer.Return.Successes = true

	cfg.Net.TLS.Config = &tls.Config{
		//Certificates: []tls.Certificate{crt},
		InsecureSkipVerify: true,
	}

	brokerList := strings.Split(cd.Brokers, ",")
	logger.Info(fmt.Sprintf("Kafka brokers: %s", strings.Join(brokerList, ", ")))

	p, err := sarama.NewSyncProducer(brokerList, cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to start Sarama producer:", err))
	}

	return &Connectors{producer: p, Name: "RealConnectors"}
}

func (r *Connectors) SendMessageSync(b []byte) error {
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	partition, offset, err := r.producer.SendMessage(&sarama.ProducerMessage{
		Topic: os.Getenv("TOPIC"),
		Value: sarama.ByteEncoder(b),
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to store your data:, %s", err))
		return err
	} else {
		// The tuple (topic, partition, offset) can be used as a unique identifier
		// for a message in a Kafka cluster.
		logger.Info(fmt.Sprintf("Your data is stored with unique identifier important/%d/%d", partition, offset))
	}

	return nil
}

func (r *Connectors) Close() {
	if err := r.producer.Close(); err != nil {
		logger.Error(fmt.Sprintf("Failed to shut down data collector cleanly", err))
	}
}
