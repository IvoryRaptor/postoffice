package iotnn

import (
	"github.com/IvoryRaptor/dragonfly"
	"gopkg.in/yaml.v2"
)

type FileWatch struct {
	dragonfly.FileWatchService
	route map[string]map[string][]string
}

func (l *FileWatch) fileChange(data []byte) error {
	l.route = map[string]map[string][]string{}
	return yaml.Unmarshal(data, &l.route)
}

func (l *FileWatch) GetMatrix(matrix string) map[string][]string {
	if _, ok := l.route[matrix]; ok {
		return l.route[matrix]
	}
	return nil
}
