package kernel

import (
	"fmt"
	"sync"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/golang/protobuf/proto"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"log"
)

type PostOffice struct {
	dragonfly.Kernel
	auth      auth.IAuthenticator
	clients   sync.Map
	zookeeper *dragonfly.Zookeeper
	mq        mq.IMQ
	redis     *dragonfly.Redis
}

func (po *PostOffice) SetFields() {
	po.auth = po.GetService("auth").(auth.IAuthenticator)
	po.clients = sync.Map{}
	po.mq = po.GetService("mq").(mq.IMQ)
	po.redis = po.GetService("redis").(*dragonfly.Redis)
	po.zookeeper = po.GetService("zookeeper").(*dragonfly.Zookeeper)
}

func (po *PostOffice) GetTopics(matrix string, action string) []string {
	m := po.zookeeper.GetChilde(matrix)
	if m == nil {
		return nil
	}
	return m.GetKeys()
}

func (po *PostOffice) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	return po.GetService("auth").(auth.IAuthenticator).Authenticate(msg)
}

func (po *PostOffice) Publish(channel *postoffice.ChannelConfig, resource string, action string, payload []byte) error {
	mes := postoffice.MQMessage{
		Source: &postoffice.Address{
			Matrix: "POSTOFFICE",
			Device: po.Get("host").(string),
		},
		Destination: &postoffice.Address{
			Matrix: channel.Matrix,
			Device: channel.DeviceName,
		},
		Resource: resource,
		Action:   action,
		Payload:  payload,
	}
	topics := po.GetTopics(channel.Matrix, resource+"."+action)
	if topics != nil {
		payload, _ := proto.Marshal(&mes)
		for _, topic := range topics {
			po.mq.Publish(topic, []byte(channel.DeviceName), payload)
		}
	} else {
		log.Printf("MISS Matrix:%s Resource:%s Action:%s", channel.Matrix, resource, action)
	}
	return nil
}

func (po *PostOffice) AddDevice(deviceName string, client postoffice.IClient) {
	po.redis.Do("HMSET", "POSTOFFICE", deviceName, po.Get("host"))
	po.clients.Store(deviceName, client)
}

func (po *PostOffice) Close(deviceName string) {
	po.redis.Do("HDEL", "POSTOFFICE", deviceName)
	po.clients.Delete(deviceName)
}

func (po *PostOffice) Arrive(msg *postoffice.MQMessage) {
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
	} else {
		log.Printf("MISS Matrix:%s Resource:%s Action:%s\n", msg.Source.Matrix, msg.Resource, msg.Action)
	}
}
