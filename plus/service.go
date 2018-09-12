package plus

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice-plus"
	mq_plus "github.com/IvoryRaptor/postoffice-plus/mqtt"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"plugin"
)

type Service struct {
	plus map[string]mq_plus.IWorkPlus
}

func (s *Service) Work(client *mqtt.Client, name string, msg *postoffice_plus.MQMessage) error {
	msgs, err := s.plus[name].Work(msg)
	if err != nil {
		return err
	}
	for _, msg := range msgs {
		client.Publish(msg)
	}
	return nil
}

func (s *Service) Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error {
	for name, config := range config {
		conf := config.(map[interface{}]interface{})
		p, err := plugin.Open("./plus/" + conf["file"].(string))
		if err != nil {
			return err
		}
		factory, err := p.Lookup("Factory")
		if err != nil {
			panic(err)
		}
		res, err := factory.(func(config map[interface{}]interface{}) (mq_plus.IWorkPlus, error))(conf)
		s.plus[name.(string)] = res
	}
	return nil
}

func (s *Service) Start() error {
	return nil
}

func (s *Service) Stop() {
}
