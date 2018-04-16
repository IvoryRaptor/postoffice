package mq

import (
	"github.com/IvoryRaptor/postoffice"
)

type IMQ interface {
	Config(kernel postoffice.IKernel, config *Config) error
	Start() error
	Stop()
}
