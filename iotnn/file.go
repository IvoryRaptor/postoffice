package iotnn

import (
	"github.com/IvoryRaptor/dragonfly"
	"gopkg.in/yaml.v2"
)

type IOTNN interface {
	dragonfly.IService
	GetMatrix(matrix string) map[string][]string
}

type FileWatch struct {
	dragonfly.FileWatchService
	route map[string]map[string][]string
}

func (l *FileWatch) FileChange(data []byte) error {
	return yaml.Unmarshal(data, &l.route)
}

func (l *FileWatch) GetMatrix(matrix string) map[string][]string {
	if _, ok := l.route[matrix]; ok {
		return l.route[matrix]
	}
	return nil
}
