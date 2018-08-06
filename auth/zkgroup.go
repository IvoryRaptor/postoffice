package auth

import (
	"sync"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
)

type ZkGroupAuth struct {
	kernel    postoffice.IPostOffice
	groups    sync.Map
	zookeeper *dragonfly.Zookeeper
}

func (z *ZkGroupAuth) Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error {
	z.kernel = kernel.(postoffice.IPostOffice)
	return nil
}

func (z *ZkGroupAuth) Start() error {
	return nil
}

func (z *ZkGroupAuth) Authenticate(block *postoffice.AuthBlock) *postoffice.ChannelConfig {
	var a postoffice.IAuthenticator
	v, ok := z.groups.Load(block.ProductKey)
	if !ok{
		z.zookeeper.GetChild(block.ProductKey)
	}else {
		a = v.(postoffice.IAuthenticator)
	}
	return a.Authenticate(block)
}

func (z *ZkGroupAuth) Stop() {
	z.kernel.RemoveService(z)
}
