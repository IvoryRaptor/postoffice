package group

//
//import (
//	"github.com/IvoryRaptor/dragonfly/mq"
//	"github.com/confluentinc/confluent-kafka-go/kafka"
//	"github.com/IvoryRaptor/dragonfly"
//	//"fmt"
//)
//
//type Service struct {
//	mq.Kafka
//	dragonfly.Zookeeper
//}
//
//func (s *Service) Send(topic string, actor []byte, payload []byte) error {
//	return s.KafkaPublish(topic, kafka.PartitionAny, actor, payload)
//}
//
//func (s *Service) Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error {
//	//kafka := kernel.GetService("mq").(* mq.Kafka)
//	//s.Topic = config["topic"].(string)
//	//err := s.KafkaConfig(kernel, config)
//	//if err != nil {
//	//	return err
//	//}
//	return nil
//}
//
//func (s *Service) Start() error {
//	s.Kafka.Start()
//	s.Zookeeper.Start()
//	return nil
//}
//
//func (s *Service) Stop() {
//	//for _, channel := range s.channels {
//	//	channel.Stop()
//	//}
//	//.RemoveService(s)
//}
