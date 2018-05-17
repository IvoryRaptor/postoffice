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
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"sync"
	"fmt"
	"github.com/golang/protobuf/proto"
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

func (kernel *Kernel) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig{
	return kernel.authenticator.Authenticate(msg)
}

func (kernel *Kernel) Publish(channel * postoffice.ChannelConfig, resource string,action string, payload []byte) error {
	mes := postoffice.MQMessage{
		Host:     kernel.GetHost(),
		Actor:    channel.DeviceName,
		Resource: resource,
		Action:   action,
		Payload:  payload,
	}
	topics, ok := kernel.GetTopics(channel.ProductKey, resource+"."+action)
	if ok {
		payload, _ := proto.Marshal(&mes)
		for _, topic := range topics {
			kernel.mq.Publish(topic, []byte(channel.DeviceName), payload)
		}
	} else {
		println(channel.ProductKey, action, "miss")
	}
	return nil
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

func (kernel *Kernel)AddDevice(deviceName string, client postoffice.IClient) {
	kernel.clients.Store(deviceName, client)
}

func (kernel *Kernel)Arrive(msg *postoffice.MQMessage) {
	val, ok := kernel.clients.Load(msg.Actor)
	if ok {
		client := val.(*mqtt.Client)
		channel := client.GetChannel()
		switch msg.Resource {
		case "system":
			switch msg.Action {
			case "close":
				client.Stop()
			default:

			}
		default:
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
	}else{
		println("123456")
	}
}

func (kernel *Kernel)GetSSL() (crt string, key string) {
	return kernel.config.SSL.Crt, kernel.config.SSL.Key
}
