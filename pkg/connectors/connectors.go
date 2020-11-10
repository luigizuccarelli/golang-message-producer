package connectors

import (
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
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
	Producer *kafka.Producer
	Logger   *simple.Logger
	Name     string
}

func NewClientConnectors(logger *simple.Logger) Clients {
	logger.Trace("Creating kafka connections for message producer")
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_BROKERS")})
	if err != nil {
		logger.Error(fmt.Sprintf("Creating kafka connections %v", err))
		panic(err)
	}

	//defer p.Close()
	return &Connectors{Producer: p, Logger: logger, Name: "RealConnectors"}
}

func (conn *Connectors) SendMessageSync(b []byte) error {
	// We are not setting a message key, which means that all messages will
	// be distributed randomly over the different partitions.
	// Delivery report handler for produced messages
	go func() {
		for e := range conn.Producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					conn.Error(fmt.Sprintf("Delivery failed message to %v", ev.TopicPartition))
				} else {
					conn.Info(fmt.Sprintf("Delivered message to %v", ev.TopicPartition))
				}
			}
		}
	}()

	topic := os.Getenv("TOPIC")
	err := conn.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          b,
	}, nil)

	//conn.Producer.Flush(100)

	return err
}

func (conn *Connectors) Close() {
	conn.Producer.Close()
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
