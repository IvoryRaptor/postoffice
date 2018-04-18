package postoffice

import (
	"net"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type IMatrix interface {
	GetTopics(action string) ([]string,bool)
}


type ChannelConfig struct{
	ClientId string
	DeviceName string
	ProductKey string
}

type IKernel interface {
	GetHost() int32
	Start() error
	AddChannel(c net.Conn) (err error)
	GetMatrix(name string) (IMatrix, bool)
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(topic string,payload []byte) error
}
