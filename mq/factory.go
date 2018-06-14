package mq

import (
	"github.com/IvoryRaptor/dragonfly"
)

type Factory struct {
}

func (f * Factory)GetName() string{
	return "mq"
}

func (f * Factory)Create(kernel dragonfly.IKernel,config map[interface {}]interface{}) (dragonfly.IService,error) {
	var mq IMQ = nil
	switch config["type"] {
	case "kafka":
		mq = &Kafka{}
		mq.Config(kernel, config)
	}
	return mq, nil
}
