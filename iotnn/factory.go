package iotnn

import (
	"github.com/IvoryRaptor/dragonfly"
)

type IOTNN interface {
	dragonfly.IService
	GetMatrix(matrix string) map[string][]string
}

type Factory struct {
}

func (f *Factory) GetName() string {
	return "iotnn"
}

func (f *Factory) Create(kernel dragonfly.IKernel, config map[interface{}]interface{}) (dragonfly.IService, error) {
	var r dragonfly.IService = nil
	switch config["type"] {
	case "file":
		f := FileWatch{}
		f.FileChange = f.fileChange
		f.Config(kernel, config)
		r = &f
	}
	return r, nil
}

var Singleton = Factory{}
