package kernel

import (
	"fmt"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/dragonfly/mq"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice-plus"
	"github.com/IvoryRaptor/postoffice-plus/mqtt/message"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/iotnn"
	pmq "github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"github.com/IvoryRaptor/postoffice/plus"
	"github.com/IvoryRaptor/postoffice/source"
	"github.com/golang/protobuf/proto"
	"github.com/surge/glog"
	"log"
	"strconv"
	"strings"
	"sync"
)

type PostOffice struct {
	dragonfly.Kernel
	auth    postoffice.IAuthenticator
	clients sync.Map
	iotnn   iotnn.IOTNN
	mq      mq.IMQ
	redis   *dragonfly.Redis
	topic   string
	plus    *plus.Service
}

func (po *PostOffice) New(hostname string) error {
	po.NewKernel("postoffice")
	po.Set("matrix", "default")
	po.Set("angler", hostname)
	err := dragonfly.Builder(
		po,
		[]dragonfly.IServiceFactory{
			&source.Singleton,
			&auth.Singleton,
			&pmq.Singleton,
			&iotnn.Singleton,
			&dragonfly.Singleton,
			&plus.Singleton,
		})
	po.auth = po.GetService("auth").(postoffice.IAuthenticator)
	po.clients = sync.Map{}
	po.mq = po.GetService("mq").(mq.IMQ)
	po.redis = po.GetService("redis").(*dragonfly.Redis)
	po.iotnn = po.GetService("iotnn").(iotnn.IOTNN)
	po.topic = fmt.Sprintf("%s_%s", po.Get("matrix"), po.Get("angler"))
	po.plus = po.GetService("work_plus").(*plus.Service)
	return err
}

func (po *PostOffice) GetTopics(matrix string, action string) []string {
	m := po.iotnn.GetMatrix(matrix)
	if m == nil {
		return nil
	}
	if _, ok := m[action]; ok {
		return m[action]
	}
	return nil
}

func (po *PostOffice) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	var block postoffice.AuthBlock
	block.ClientId = string(msg.ClientId())
	block.Password = string(msg.Password())
	block.Username = string(msg.Username())

	if strings.Index(block.ClientId, "|") > 0 {
		sp := strings.Split(block.Username, "&")
		if len(sp) == 2 {
			block.DeviceName = sp[0]
			block.ProductKey = sp[1]

			sp = strings.Split(string(msg.ClientId()), "|")
			if len(sp) != 3 {
				return nil
			}
			block.ClientId = sp[0]
			sp = strings.Split(sp[1], ",")
			block.Params = map[string]string{
				"clientId":   block.ClientId,
				"productKey": block.ProductKey,
				"deviceName": block.DeviceName,
			}
			block.Keys = []string{"productKey", "deviceName", "clientId"}
			var err error
			for _, i := range sp {
				v := strings.Split(i, "=")
				switch v[0] {
				case "securemode":
					block.SecureMode, err = strconv.Atoi(v[1])
					if err != nil {
						log.Printf("Unknown securemode %s", v[1])
						return nil
					}
				case "signmethod":
					block.SignMethod = v[1]
				case "timestamp":
					block.Keys = append(block.Keys, v[0])
					block.Params[v[0]] = v[1]
				}
			}
		}
	}
	res := po.GetService("auth").(postoffice.IAuthenticator).Authenticate(&block)
	if res == nil {
		log.Printf("auth fail %s %s %s", block.ClientId, block.Username, block.Password)
	}
	return res
}

func (po *PostOffice) Publish(channel *postoffice.ChannelConfig, resource string, action string, payload []byte) error {
	mes := postoffice_plus.MQMessage{
		Caller:   []byte(po.topic),
		Matrix:   channel.Matrix,
		Device:   channel.DeviceName,
		Resource: resource,
		Action:   action,
		Payload:  payload,
	}
	topics := po.GetTopics(channel.Matrix, resource+"."+action)
	if topics != nil {
		payload, _ := proto.Marshal(&mes)
		for _, topic := range topics {
			log.Printf("topics [%s]", topic)
			if len(topic) > 5 && topic[0:5] == "plus." {

				val, ok := po.clients.Load(channel.Matrix + "/" + channel.DeviceName)
				if ok {
					client := val.(*mqtt.Client)
					po.plus.Work(client, topic[6:], &mes)
				}
			} else {
				po.mq.Publish(topic, []byte(channel.DeviceName), payload)
			}
			switch topic {
			case "reply":
				val, ok := po.clients.Load(channel.Matrix + "/" + channel.DeviceName)
				if ok {
					client := val.(*mqtt.Client)
					pus := message.NewPublishMessage()
					topic := fmt.Sprintf(
						"%s/%s/%s/%s",
						channel.Matrix,
						channel.DeviceName,
						resource,
						action)
					pus.SetTopic([]byte(topic))
					pus.SetPayload(payload)
					client.Publish(pus)
				}
			default:

			}
		}
	} else {
		log.Printf("MISS Matrix:%s Resource:%s Action:%s", channel.Matrix, resource, action)
	}
	return nil
}

func (po *PostOffice) AddDevice(deviceName string, client postoffice.IClient) {
	glog.Info("AddDevice")
	po.redis.Do("HMSET", "POSTOFFICE", deviceName, po.topic)
	po.clients.Store(deviceName, client)
}

func (po *PostOffice) Close(deviceName string) {
	po.redis.Do("HDEL", "POSTOFFICE", deviceName)
	po.clients.Delete(deviceName)
}

func (po *PostOffice) Arrive(data []byte) {
	msg := postoffice_plus.MQMessage{}
	err := proto.Unmarshal(data, &msg)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("%s=>%s/%s|%s.%s", msg.Caller, msg.Matrix, msg.Device, msg.Action, msg.Resource)
	val, ok := po.clients.Load(msg.Matrix + "/" + msg.Device)
	if ok {
		client := val.(postoffice.IClient)
		switch msg.Resource {
		case "system":
			switch msg.Action {
			case "close":
				client.Stop()
			default:

			}
		default:
			client.Send(&msg)
		}
	} else {
		log.Printf("MISS Matrix:%s Resource:%s Action:%s\n", msg.Matrix, msg.Resource, msg.Action)
	}
}
