package postoffice

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/dragonfly/mq"
	"github.com/IvoryRaptor/postoffice-plus"
	"github.com/IvoryRaptor/postoffice-plus/mqtt/message"
)

type ChannelConfig struct {
	DeviceName string
	Matrix     string
}

type AuthBlock struct {
	ClientId   string
	DeviceName string
	ProductKey string
	SecureMode int
	SignMethod string
	Keys       []string
	Params     map[string]string
	Username   string
	Password   string
}

type IAuthenticator interface {
	dragonfly.IService
	Authenticate(block *AuthBlock) *ChannelConfig
}

type IClient interface {
	Send(msg *postoffice_plus.MQMessage) error
	Stop()
}

type IPostOffice interface {
	mq.IArrive
	GetTopics(matrix string, action string) []string
	Authenticate(msg *message.ConnectMessage) *ChannelConfig
	Publish(channel *ChannelConfig, resource string, action string, payload []byte) error
	AddDevice(device string, client IClient)
	Close(device string)
}
