package mq

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/protobuf/proto"
	"fmt"
	"os"
	"log"
	"github.com/IvoryRaptor/dragonfly"
)

type Kafka struct {
	kernel   postoffice.IPostOffice
	producer *kafka.Producer
	consumer *kafka.Consumer
}

func (k * Kafka)GetName() string{
	return "mq"
}

func (k * Kafka)Publish(topic string,actor []byte,payload []byte) error {
	deliveryChan := make(chan kafka.Event)
	err := k.producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Key:            actor,
			Value:          payload,
		},
		deliveryChan)
	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
	if err != nil {
		fmt.Println(topic, err.Error())
		return err
	}
	return nil
}

func (k * Kafka)Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error{
	k.kernel = kernel.(postoffice.IPostOffice)
	var err error = nil
	host := fmt.Sprintf("%s:%d",config["host"],config["port"])
	k.producer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": host})
	if err != nil {
		return err
	}

	t := kafka.ConfigMap{
		"bootstrap.servers":    host,
		"group.id":             "PostOffice",
		"session.timeout.ms":   6000,
		"default.topic.config": kafka.ConfigMap{"auto.offset.reset": "earliest"}}
	k.consumer, err = kafka.NewConsumer(&t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	return nil
}

func (k * Kafka)Start() error {
	log.Printf("mq start")
	err := k.consumer.SubscribeTopics([]string{fmt.Sprintf("postoffice-%s", k.kernel.Get("host"))}, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create consumer: %s\n", err)
		os.Exit(1)
	}
	go func() {
		for true {
			ev := k.consumer.Poll(100)
			if ev == nil {
				continue
			}
			switch e := ev.(type) {
			case *kafka.Message:
				msg := postoffice.MQMessage{}
				err := proto.Unmarshal(e.Value, &msg)
				if err != nil {
					log.Println(err.Error())
				}
				k.kernel.Arrive(&msg)
				log.Printf("%s.%s=>%s/%s", msg.Resource, msg.Action, msg.Source.Matrix,msg.Source.Device)
			case kafka.PartitionEOF:
				fmt.Printf("%% Reached %v\n", e)
			case kafka.Error:
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			default:
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}()
	//mq.consumer.Close()
	return nil
}

func (k * Kafka)Stop(){
	k.kernel.RemoveService(k)
}
