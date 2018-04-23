package postoffice

import (
	"net"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type ChannelConfig struct{
	ClientId string
	DeviceName string
	ProductKey string
}

type IClient interface {
	Publish(msg *message.PublishMessage) error
}

type IKernel interface {
	GetHost() int32
	Start() error
	AddChannel(c net.Conn) (err error)
	GetTopics(matrix string, action string) ([]string, bool)
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(topic string, payload []byte) error
	AddClient(clientId string, client IClient)
	Arrive(msg *MQMessage)
}
