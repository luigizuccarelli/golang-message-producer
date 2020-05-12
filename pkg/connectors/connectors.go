package connectors

import (
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/microlib/simple"
)

type Clients interface {
	Error(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Trace(string, ...interface{})
	SendMessageSync(body []byte) error
	Close()
}

type Connectors struct {
	Producer sarama.SyncProducer
	Logger   *simple.Logger
	Name     string
}

func NewClientConnectors(logger *simple.Logger) Clients {

	logger.Trace("Creating Kafka message producer")

	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	cfg.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	cfg.Producer.Return.Successes = true

	cfg.Net.TLS.Config = &tls.Config{
		//Certificates: []tls.Certificate{crt},
		InsecureSkipVerify: true,
	}

	brokerList := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	logger.Info(fmt.Sprintf("Kafka brokers: %s", strings.Join(brokerList, ", ")))

	p, err := sarama.NewSyncProducer(brokerList, cfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to start Sarama producer : %v", err))
	}

	return &Connectors{Producer: p, Logger: logger, Name: "RealConnectors"}
}

func (conn *Connectors) SendMessageSync(b []byte) error {
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	partition, offset, err := conn.Producer.SendMessage(&sarama.ProducerMessage{
		Topic: os.Getenv("TOPIC"),
		Value: sarama.ByteEncoder(b),
	})

	if err != nil {
		conn.Error(fmt.Sprintf("Failed to store your data: %v", err))
		return err
	} else {
		// The tuple (topic, partition, offset) can be used as a unique identifier
		// for a message in a Kafka cluster.
		conn.Debug(fmt.Sprintf("Your data is stored with unique identifier important /%d/%d", partition, offset))
	}

	return nil
}

func (conn *Connectors) Close() {
	if err := conn.Producer.Close(); err != nil {
		conn.Error(fmt.Sprintf("Failed to shut down data collector cleanly %v", err))
	}
}

func (conn *Connectors) Error(msg string, val ...interface{}) {
	conn.Logger.Error(fmt.Sprintf(msg, val...))
}

func (conn *Connectors) Info(msg string, val ...interface{}) {
	conn.Logger.Info(fmt.Sprintf(msg, val...))
}

func (conn *Connectors) Debug(msg string, val ...interface{}) {
	conn.Logger.Debug(fmt.Sprintf(msg, val...))
}

func (conn *Connectors) Trace(msg string, val ...interface{}) {
	conn.Logger.Trace(fmt.Sprintf(msg, val...))
}
