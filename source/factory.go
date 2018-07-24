package source

import (
	"github.com/IvoryRaptor/dragonfly"
	"fmt"
	"errors"
)

type Factory struct {
}

func (f *Factory) GetName() string {
	return "source"
}

func (f *Factory) Create(kernel dragonfly.IKernel, config map[interface{}]interface{}) (dragonfly.IService, error) {
	service := Service{}
	service.Config(kernel, config)

	channels := config["channels"].([]interface{})
	service.channels = make([]IChannel, len(channels))
	for i, item := range channels {
		channel := item.(map[interface{}]interface{})
		switch channel["type"].(string) {
		case "websocket":
			service.channels[i] = &WebSocketChannel{}
		case "mqtt":
			service.channels[i] = &MQTTChannel{}
		default:
			return nil, errors.New(fmt.Sprintf("unknown source type %s", channel["type"]))
		}
		err := service.channels[i].Config(&service, channel)
		if err != nil {
			return nil, err
		}
	}
	return &service, nil
}
