package kernel

import (
	"github.com/IvoryRaptor/postoffice/matrix"
	"github.com/IvoryRaptor/postoffice/source"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice/auth"
)

type SSLConfig struct {
	Crt string `yaml:"crt"`
	Key string `yaml:"key"`
}

type Config struct {
	KeepAlive int `yaml:"keepAlive"`
	// The number of seconds to wait for the CONNECT message before disconnecting.
	// If not set then default to 2 seconds.
	ConnectTimeout int `yaml:"connectTimeout"`
	// The number of seconds to wait for any ACK messages before failing.
	// If not set then default to 20 seconds.
	AckTimeout int `yaml:"ackTimeout"`
	// The number of times to retry sending a packet if ACK is not received.
	// If no set then default to 3 retries.
	TimeoutRetries int             `yaml:"timeoutRetries"`
	Matrix         matrix.Config   `yaml:"matrix"`
	SSL            SSLConfig       `yaml:"ssl"`
	Source         []source.Config `yaml:"source"`
	MQ             mq.Config       `yaml:"mq"`
	Auth           auth.Config     `yaml:"auth"`
}
