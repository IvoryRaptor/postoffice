package postoffice

import (
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type ChannelConfig struct{
	ClientId string
	DeviceName []byte
	ProductKey string
}

type IClient interface {
	Publish(msg *message.PublishMessage) error
}

type IKernel interface {
	GetHost() int32
	Start() error
	GetTopics(matrix string, action string) ([]string, bool)
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(channel * ChannelConfig, resource string,action string, payload []byte) error

	AddClient(clientId string, client IClient)
	Arrive(msg *MQMessage)
	GetSSL() (crt string, key string)
}
