package postoffice

import (
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/dragonfly/mq"
)

type ChannelConfig struct {
	DeviceName string
	Matrix     string
	Token      string
}

type IClient interface {
	Publish(msg *message.PublishMessage) error
}

type IPostOffice interface {
	mq.IArrive
	GetTopics(matrix string, action string) []string
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(channel *ChannelConfig, resource string, action string, payload []byte) error
	AddDevice(device string, client IClient)
	Close(device string)
}
