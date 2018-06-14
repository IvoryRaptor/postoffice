package source

import (
	"github.com/IvoryRaptor/dragonfly"
)

type Service struct {
	kernel   dragonfly.IKernel
	name     string
	run      bool
	channels []IChannel
	crt      string
	key      string
	config map[interface {}]interface{}
}

func (s * Service)GetName() string{
	return "source"
}

func (s * Service)Start() error {
	for _, channel := range s.channels {
		err := channel.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s * Service)Stop(){
	for _, channel := range s.channels {
		channel.Stop()
	}
	s.kernel.RemoveService(s)
}

func (s * Service)Config(kernel dragonfly.IKernel, config map[interface {}]interface{}) error {
	s.kernel = kernel
	s.config = config
	return nil
}
