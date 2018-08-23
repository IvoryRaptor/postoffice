package auth

import (
	"errors"
	"fmt"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
)

var (
	ErrAuthFailure = errors.New("auth: Authentication failure")
)

type Factory struct {
}

func (f *Factory) GetName() string {
	return "auth"
}

func (f *Factory) Create(kernel dragonfly.IKernel, config map[interface{}]interface{}) (dragonfly.IService, error) {
	var result postoffice.IAuthenticator
	switch config["type"] {
	case "mongodb":
		result = &MongoAuth{}
	case "mock":
		result = &Mock{}
	case "redis":
		result = &RedisAuth{}
	case "oauth":
		result = &OAuth{}
	case "group":
		g := GroupAuth{}
		g.FileChange = g.fileChange
		result = &g
	default:
		return nil, errors.New(fmt.Sprintf("unknown auth type %s", config["type"]))
	}
	result.Config(kernel, config)
	return result, nil
}

var Singleton = Factory{}
