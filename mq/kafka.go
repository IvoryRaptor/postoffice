package mq

import "github.com/IvoryRaptor/postoffice"

type Kafka struct {

}

func (k * Kafka)Publish(topic string,payload []byte) error {
	return nil
}

func (k * Kafka)Config(kernel postoffice.IKernel, config *Config) error{
	return nil
}

func (k * Kafka)Start() error{
	return nil
}

func (k * Kafka)Stop(){
}
