package auth

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
	"gopkg.in/yaml.v2"
)

type GroupAuth struct {
	dragonfly.FileWatchService
	kernel postoffice.IPostOffice
	groups map[string]dragonfly.IService
}

func (g *GroupAuth) FileChange(data []byte) error {
	var configs []map[interface{}]interface{}
	if err := yaml.Unmarshal(data, configs); err != nil {
		return err
	}
	n := map[string]dragonfly.IService{}
	for _, v := range configs {
		if s, err := Singleton.Create(g.kernel, v); err == nil {
			n[v["matrix"].(string)] = s
			s.Start()
		}
	}
	o := g.groups
	g.groups = n
	for _, i := range o {
		i.Stop()
	}
	return nil
}

func (g *GroupAuth) Authenticate(block *postoffice.AuthBlock) *postoffice.ChannelConfig {
	var a postoffice.IAuthenticator
	if v, ok := g.groups[block.ProductKey]; ok {
		a = v.(postoffice.IAuthenticator)
	} else {
		return nil
	}
	return a.Authenticate(block)
}
