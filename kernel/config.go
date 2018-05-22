package kernel

import (
	"github.com/IvoryRaptor/postoffice/matrix"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/ssl"
)

type RedisConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	Auth   auth.Config              `yaml:"auth"`
	Matrix matrix.Config            `yaml:"matrix"`
	MQ     mq.Config                `yaml:"mq"`
	Source []map[string]interface{} `yaml:"source"`
	SSL    ssl.Config               `yaml:"ssl"`
	Redis RedisConfig				`yaml:"redis"`
}
