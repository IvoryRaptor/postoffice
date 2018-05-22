package postoffice

import (
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type ChannelConfig struct{
	DeviceName string
	ProductKey string
	Token string
}

type IClient interface {
	Publish(msg *message.PublishMessage) error
}

type IKernel interface {
	GetHost() string
	Start() error
	GetTopics(matrix string, action string) ([]string, bool)
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(channel *ChannelConfig, resource string, action string, payload []byte) error

	AddDevice(device string, client IClient)
	Arrive(msg *MQMessage)
	GetSSL() (crt string, key string)
	Close(device string)
}
