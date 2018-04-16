package kernel

import (
	"github.com/IvoryRaptor/postoffice/matrix"
	"github.com/IvoryRaptor/postoffice/source"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/ssl"
	"github.com/IvoryRaptor/postoffice/mqtt"
)


type Config struct {
	Auth   auth.Config     `yaml:"auth"`
	Matrix matrix.Config   `yaml:"matrix"`
	MQ     mq.Config       `yaml:"mq"`
	MQTT   mqtt.Config     `yaml:"mqtt"`
	Source []source.Config `yaml:"source"`
	SSL    ssl.Config      `yaml:"ssl"`
}
