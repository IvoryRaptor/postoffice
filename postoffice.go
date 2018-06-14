package postoffice

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type ChannelConfig struct{
	DeviceName string
	Matrix     string
	Token      string
}

type IClient interface {
	Publish(msg *message.PublishMessage) error
}

type IPostOffice interface {
	dragonfly.IKernel
	GetTopics(matrix string, action string) ([]string, bool)
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(channel *ChannelConfig, resource string, action string, payload []byte) error
	AddDevice(device string, client IClient)
	Arrive(msg *MQMessage)
	Close(device string)
}
