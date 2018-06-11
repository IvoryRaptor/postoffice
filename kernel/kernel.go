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

type PostOffice struct {
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

func (po *PostOffice)IsRun() bool {
	return po.run
}

func (po *PostOffice)GetHost() string{
	return po.host
}

func (po *PostOffice)GetTopics(matrix string, action string) ([]string, bool){
	m,ok := po.matrixManger.GetMatrix(matrix)
	if !ok{
		return nil,false
	}
	return m.GetTopics(action)
}

func (po *PostOffice) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig{
	return po.authenticator.Authenticate(msg)
}

func (po *PostOffice) Publish(channel * postoffice.ChannelConfig, resource string,action string, payload []byte) error {
	mes := postoffice.MQMessage{
		Source: &postoffice.Address{
			Matrix: "POSTOFFICE",
			Device: po.GetHost(),
		},
		Destination:&postoffice.Address{
			Matrix:channel.Matrix,
			Device:channel.DeviceName,
		},
		Resource: resource,
		Action:   action,
		Payload:  payload,
	}
	topics, ok := po.GetTopics(channel.Matrix, resource+"."+action)
	if ok {
		payload, _ := proto.Marshal(&mes)
		for _, topic := range topics {
			po.mq.Publish(topic, []byte(channel.DeviceName), payload)
		}
	} else {
		println(channel.Matrix, action, "miss")
	}
	return nil
}

func (po *PostOffice) WaitStop() {
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	po.Stop()
	stopChan <- struct{}{}
	os.Exit(0)
}

func (po *PostOffice)AddDevice(deviceName string, client postoffice.IClient) {
	po.redisMutex.Lock()
	po.redis.Do("HMSET", "POSTOFFICE", deviceName, po.host)
	po.clients.Store(deviceName, client)
	po.redisMutex.Unlock()
}

func (po *PostOffice)Close(deviceName string){
	po.redisMutex.Lock()
	po.redis.Do("HDEL", "POSTOFFICE", deviceName)
	po.clients.Delete(deviceName)
	po.redisMutex.Unlock()
}

func (po *PostOffice)Arrive(msg *postoffice.MQMessage) {
	val, ok := po.clients.Load(msg.Source.Matrix + "/" + msg.Source.Device)
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

func (po *PostOffice)GetSSL() (crt string, key string) {
	return po.config.SSL.Crt, po.config.SSL.Key
}
