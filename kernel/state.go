package kernel

import (
	"regexp"
	"strconv"
	"github.com/IvoryRaptor/postoffice/source"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"log"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/mq"
	"errors"
	"fmt"
)

func (kernel *Kernel)Config(hostname string)error {
	var err error

	log.Println("Config HostName:" + hostname)
	//Get kubernetes hostname
	reg := regexp.MustCompile(`(\d+)`)
	host, err := strconv.Atoi(reg.FindString(hostname))
	if err != nil {
		return err
	}
	kernel.host = int32(host)
	//Load Config
	log.Println("Load Config File", kernel.ConfigFile)
	data, err := ioutil.ReadFile(kernel.ConfigFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal([]byte(data), &kernel.config)
	if err != nil {
		return err
	}

	kernel.matrixManger.Config(kernel, &kernel.config.Matrix)

	log.Println("Config MQ")
	switch kernel.config.MQ.Type {
	case "kafka":
		kernel.mq = &mq.Kafka{}
	default:
		return errors.New(fmt.Sprintf("unknown mq type %s", kernel.config.MQ.Type))
	}
	err = kernel.mq.Config(kernel, &kernel.config.MQ)
	if err != nil {
		return err
	}

	//set source config
	log.Println("Config Source")
	kernel.source = make([]source.ISource, len(kernel.config.Source))

	for i, item := range kernel.config.Source {
		switch item["type"].(string) {
		case "websocket":
			kernel.source[i] = &source.WebSocketSource{}
		case "mqtt":
			kernel.source[i] = &source.MQTTSource{}
		case "coap":
			kernel.source[i] = &source.CoapSource{}
		default:
			log.Fatalf("unknow source type %s", item["type"].(string))
		}
		err = kernel.source[i].Config(kernel, item)
		if err != nil {
			return err
		}
	}
	//kernel.authenticator = &auth.Mock{}
	kernel.authenticator = &auth.Mongo{}
	kernel.authenticator.Config(kernel, &kernel.config.Auth)
	return nil
}

func (kernel *Kernel) Start() error {
	var err error
	kernel.run = true
	log.Println("Start MQ")
	err = kernel.mq.Start()
	if err != nil {
		return err
	}

	log.Println("Start Matrix Manager")
	err = kernel.matrixManger.Start()
	if err != nil {
		return err
	}

	log.Println("Start Source")
	for _, item := range kernel.source {
		err = item.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (kernel *Kernel) Stop() {
	kernel.run = false
	for _, item := range kernel.source {
		item.Stop()
	}
	kernel.mq.Stop()
	kernel.matrixManger.Stop()
}
