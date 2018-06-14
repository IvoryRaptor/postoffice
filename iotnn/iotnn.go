package iotnn

import "github.com/IvoryRaptor/dragonfly"

type IMatrix interface {
	GetTopics(action string) ([]string,bool)
}

type IIotNN interface {
	dragonfly.IService
	GetMatrix(matrix string) (IMatrix,bool)
}
