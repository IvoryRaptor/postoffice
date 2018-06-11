package kernel

import (
	"regexp"
	"github.com/IvoryRaptor/postoffice/source"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"log"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/mq"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
)

func (po *PostOffice)Config(hostname string)error {
	var err error

	log.Println("Config HostName:" + hostname)
	//Get kubernetes hostname
	reg := regexp.MustCompile(`(\d+)`)
	po.host = reg.FindString(hostname)
	//Load Config
	log.Println("Load Config File", po.ConfigFile)
	data, err := ioutil.ReadFile(po.ConfigFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(data), &po.config)
	if err != nil {
		return err
	}

	po.iotnnManger.Config(po, &po.config.Matrix)

	log.Println("Config MQ")
	switch po.config.MQ.Type {
	case "kafka":
		po.mq = &mq.Kafka{}
	default:
		return errors.New(fmt.Sprintf("unknown mq type %s", po.config.MQ.Type))
	}
	err = po.mq.Config(po, &po.config.MQ)
	if err != nil {
		return err
	}

	//set source config
	log.Println("Config Source")
	po.source = make([]source.ISource, len(po.config.Source))

	for i, item := range po.config.Source {
		switch item["type"].(string) {
		case "websocket":
			po.source[i] = &source.WebSocketSource{}
		case "mqtt":
			po.source[i] = &source.MQTTSource{}
		case "coap":
			po.source[i] = &source.CoapSource{}
		default:
			log.Fatalf("unknow source type %s", item["type"].(string))
		}
		err = po.source[i].Config(po, item)
		if err != nil {
			return err
		}
	}
	//po.authenticator = &auth.Mock{}
	po.authenticator = &auth.Mongo{}
	po.authenticator.Config(po, &po.config.Auth)
	return nil
}

func (po *PostOffice) Start() error {
	var err error
	po.run = true
	c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", po.config.Redis.Host, po.config.Redis.Port))
	if err != nil {
		return err
	}
	po.redis = c

	log.Println("Start MQ")
	err = po.mq.Start()
	if err != nil {
		return err
	}

	log.Println("Start Matrix Manager")
	err = po.iotnnManger.Start()
	if err != nil {
		return err
	}

	log.Println("Start Source")
	for _, item := range po.source {
		err = item.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (po *PostOffice) Stop() {
	po.redis.Close()
	po.run = false
	for _, item := range po.source {
		item.Stop()
	}
	po.mq.Stop()
	po.iotnnManger.Stop()
}
