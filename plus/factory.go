package plus

import (
	"github.com/IvoryRaptor/dragonfly"
)

type Factory struct {
}

func (f *Factory) GetName() string {
	return "work_plus"
}

func (f *Factory) Create(kernel dragonfly.IKernel, config map[interface{}]interface{}) (dragonfly.IService, error) {
	service := Service{}
	service.Config(kernel, config)
	return &service, nil
}

var Singleton = Factory{}
