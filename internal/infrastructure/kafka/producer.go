package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

// Producer обертка над Sarama
type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer создает подключение
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	// Ждем подтверждения от Kafka, что сообщение записано
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

// SendMessage отправляет любой struct как JSON
func (p *Producer) SendMessage(key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshalling error: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key), // Key нужен, чтобы сообщения одного юзера шли в одну партицию
		Value: sarama.ByteEncoder(bytes),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("kafka send error: %w", err)
	}

	log.Printf("[Kafka] Sent to partition %d offset %d", partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
