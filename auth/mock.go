package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type Mock struct {
	kernel postoffice.IKernel
}

func (a *Mock) Config(kernel postoffice.IKernel,config *Config) error{
	return nil
}

func (a *Mock) Start() error{
	return nil
}

func (a *Mock) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	config := postoffice.ChannelConfig{
		ClientId:   string(msg.ClientId()),
		DeviceName: string(msg.Username()),
		ProductKey: string(msg.Password()),
	}
	return &config
}
