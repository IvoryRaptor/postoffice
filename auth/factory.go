package auth

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"errors"
	"fmt"
)

var (
	ErrAuthFailure          = errors.New("auth: Authentication failure")
)

type IAuthenticator interface {
	dragonfly.IService
	Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig
}

type Factory struct {
}

func (f * Factory)GetName() string{
	return "auth"
}

func (f * Factory)Create(kernel dragonfly.IKernel,config map[interface {}]interface{}) (dragonfly.IService,error) {
	var result IAuthenticator
	switch config["type"] {
	case "mongodb":
		result = &MongoAuth{}
	case "mock":
		result = &Mock{}
	case "redis":
		result = &RedisAuth{}
	default:
		return nil, errors.New(fmt.Sprintf("unknown auth type %s", config["type"]))
	}
	result.Config(kernel, config)
	return result, nil
}
