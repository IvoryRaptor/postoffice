package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/dragonfly"
)

type Mock struct {
	kernel postoffice.IPostOffice
}

func (m *Mock) Config(kernel dragonfly.IKernel, config map[interface {}]interface{}) error{
	m.kernel = kernel.(postoffice.IPostOffice)
	return nil
}

func (m *Mock) Start() error{
	return nil
}

func (m *Mock) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	config := postoffice.ChannelConfig{
		//ClientId:   string(msg.ClientId()),
		DeviceName: string(msg.Username()),
		Matrix:     string(msg.Password()),
	}
	return &config
}

func (m *Mock) Stop(){
	m.kernel.RemoveService(m)
}