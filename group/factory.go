package group

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/dragonfly/mq"
)

type Factory struct {
}

func (f *Factory) GetName() string {
	return "mq"
}

func (f *Factory) Create(kernel dragonfly.IKernel, config map[interface{}]interface{}) (dragonfly.IService, error) {
	var m mq.IMQ = nil
	switch config["type"] {
	case "kafka":
		m = &Service{}
		m.Config(kernel, config)
	}
	return m, nil
}
