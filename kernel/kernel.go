package kernel

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/IvoryRaptor/postoffice/source"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/matrix"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"net"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"sync"
	"fmt"
)

type Kernel struct {
	host          int32
	ConfigFile    string
	run           bool
	source        []source.ISource
	authenticator auth.IAuthenticator
	matrixManger  matrix.Manager
	config        Config
	mq            mq.IMQ
	mqtt          mqtt.MQTT
	clients       sync.Map
}

func (kernel *Kernel)IsRun() bool {
	return kernel.run
}

func (kernel *Kernel)GetHost() int32{
	return kernel.host
}

func (kernel *Kernel)GetTopics(matrix string, action string) ([]string, bool){
	m,ok :=kernel.matrixManger.GetMatrix(matrix)
	if !ok{
		return nil,false
	}
	return m.GetTopics(action)
}

func (kernel *Kernel) AddChannel(c net.Conn) (err error){
	go kernel.mqtt.AddChannel(c)
	return nil
}

func (kernel *Kernel) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig{
	return kernel.authenticator.Authenticate(msg)
}

func (kernel *Kernel) Publish(topic string,payload []byte) error {
	return kernel.mq.Publish(topic, payload)
}

func (kernel *Kernel) WaitStop() {
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	kernel.Stop()
	stopChan <- struct{}{}
	os.Exit(0)
}

func (kernel *Kernel)AddClient(clientId string, client postoffice.IClient) {
	kernel.clients.Store(clientId, client)
}

func (kernel *Kernel)Arrive(msg *postoffice.MQMessage) {
	val, ok := kernel.clients.Load(msg.Actor)
	if ok {
		client := val.(*mqtt.Client)
		channel := client.GetChannel()
		pus := message.NewPublishMessage()
		topic := fmt.Sprintf(
			"%s/%s/%s/%s",
			channel.ProductKey,
			channel.DeviceName,
			msg.Resource,
			msg.Action)
		pus.SetTopic([]byte(topic))
		pus.SetPayload(msg.Payload)
		client.Publish(pus)
	}
}

func (kernel *Kernel)GetSSL() (crt string, key string) {
	return kernel.config.SSL.Crt, kernel.config.SSL.Key
}
