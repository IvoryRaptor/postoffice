package postoffice

import (
	"net"
	"sync"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type Matrix struct {
	Name          string
	Authorization string
	Action        sync.Map
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
	GetMatrix(name string) (*Matrix, bool)
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(topic string,payload []byte) error
}
