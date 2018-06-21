package mq

import (
	"github.com/IvoryRaptor/dragonfly/mq"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/IvoryRaptor/dragonfly"
)

type Kafka struct {
	mq.Kafka
}

func (k *Kafka) Publish(topic string, actor []byte, payload []byte) error {
	return k.KafkaPublish(topic, kafka.PartitionAny, actor, payload)
}

func (k *Kafka) Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error {
	err := k.KafkaConfig(kernel, config)
	if err!=nil{
		return err
	}
	return nil
}

