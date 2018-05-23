package kernel

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/IvoryRaptor/postoffice/source"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/matrix"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"sync"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/garyburd/redigo/redis"
)

type Kernel struct {
	host          string
	ConfigFile    string
	run           bool
	source        []source.ISource
	authenticator auth.IAuthenticator
	matrixManger  matrix.Manager
	config        Config
	mq            mq.IMQ
	clients       sync.Map
	redis         redis.Conn
	redisMutex    sync.Mutex
}

func (kernel *Kernel)IsRun() bool {
	return kernel.run
}

func (kernel *Kernel)GetHost() string{
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
		Source: &postoffice.Address{
			Matrix: "POSTOFFICE",
			Device: kernel.GetHost(),
		},
		Destination:&postoffice.Address{
			Matrix:channel.Matrix,
			Device:channel.DeviceName,
		},
		Resource: resource,
		Action:   action,
		Payload:  payload,
	}
	topics, ok := kernel.GetTopics(channel.Matrix, resource+"."+action)
	if ok {
		payload, _ := proto.Marshal(&mes)
		for _, topic := range topics {
			kernel.mq.Publish(topic, []byte(channel.DeviceName), payload)
		}
	} else {
		println(channel.Matrix, action, "miss")
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
	kernel.redisMutex.Lock()
	kernel.redis.Do("HMSET", "POSTOFFICE", deviceName, kernel.host)
	kernel.clients.Store(deviceName, client)
	kernel.redisMutex.Unlock()
}

func (kernel *Kernel)Close(deviceName string){
	kernel.redisMutex.Lock()
	kernel.redis.Do("HDEL", "POSTOFFICE", deviceName)
	kernel.clients.Delete(deviceName)
	kernel.redisMutex.Unlock()
}

func (kernel *Kernel)Arrive(msg *postoffice.MQMessage) {
	val, ok := kernel.clients.Load(msg.Source.Matrix + "/" + msg.Source.Device)
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
				channel.Matrix,
				channel.DeviceName,
				msg.Resource,
				msg.Action)
			pus.SetTopic([]byte(topic))
			pus.SetPayload(msg.Payload)
			client.Publish(pus)
		}
	}else{
		println("miss")
	}
}

func (kernel *Kernel)GetSSL() (crt string, key string) {
	return kernel.config.SSL.Crt, kernel.config.SSL.Key
}
