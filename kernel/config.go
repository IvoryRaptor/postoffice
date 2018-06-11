package kernel

import (
	"github.com/IvoryRaptor/postoffice/iotnn"
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
	Matrix iotnn.Config             `yaml:"iotnn"`
	MQ     mq.Config                `yaml:"mq"`
	Source []map[string]interface{} `yaml:"source"`
	SSL    ssl.Config               `yaml:"ssl"`
	Redis  RedisConfig              `yaml:"redis"`
}
