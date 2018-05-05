package source

import "github.com/IvoryRaptor/postoffice"


type ISource interface {
	Config(kernel postoffice.IKernel, config map[string]interface{}) error
	Start() error
	Stop()
}
