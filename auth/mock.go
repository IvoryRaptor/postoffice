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

func (a *Mock) Authenticate(msg *message.ConnectMessage) error{
	return nil
}
