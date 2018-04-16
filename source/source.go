package source

import "github.com/IvoryRaptor/postoffice"

type Config struct {
	Type string		`yaml:"type"`
	SSL bool		`yaml:"ssl"`
	Port int		`yaml:"port"`
}

type ISource interface{
	Config(kernel postoffice.IKernel, config Config,crt string,key string) error
	Start() error
	Stop()
}
