package mq

import (
	"github.com/IvoryRaptor/dragonfly"
)

type IMQ interface {
	dragonfly.IService
	Publish(topic string, actor []byte, payload []byte) error
}
