package iotnn

import (
	"github.com/IvoryRaptor/dragonfly"
)

type Factory struct {
}

func (f *Factory) GetName() string {
	return "iotnn"
}

func (f *Factory) Create(kernel dragonfly.IKernel, config map[interface{}]interface{}) (dragonfly.IService, error) {
	var r dragonfly.IService = nil
	switch config["type"] {
	case "file":
		r = &FileWatch{}
		r.Config(kernel, config)
	}
	return r, nil
}

var Singleton = Factory{}
