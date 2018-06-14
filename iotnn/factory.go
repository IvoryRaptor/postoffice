package iotnn

import (
	"github.com/IvoryRaptor/dragonfly"
	"fmt"
	"errors"
)

type Factory struct {
}

func (f * Factory)GetName() string{
	return "iotnn"
}

func (f * Factory)Create(kernel dragonfly.IKernel,config map[interface {}]interface{}) (dragonfly.IService,error) {
	var result IIotNN
	switch config["type"] {
	case "zookeeper":
		result = &ZkIotNN{}
	default:
		return nil, errors.New(fmt.Sprintf("unknown iotnn type %s", config["type"]))
	}
	result.Config(kernel, config)
	return result, nil
}
