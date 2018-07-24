package source

import (
	"net"
	"fmt"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"log"
	"crypto/tls"
	"github.com/IvoryRaptor/postoffice"
)

type IChannel interface {
	Config(service *Service, config map[interface{}]interface{}) error
	Start() error
	Stop()
}

type MQTTChannel struct {
	mqtt.MQTT
	ln      net.Listener
	port    string
	ssl     bool
	service *Service
}

func (s *MQTTChannel) Config(service *Service, config map[interface{}]interface{}) error {
	s.ssl = config["ssl"].(bool)
	s.port = fmt.Sprintf(":%d", config["port"].(int))
	s.Kernel = service.kernel.(postoffice.IPostOffice)
	s.KeepAlive = service.config["keepAlive"].(int)
	s.ConnectTimeout = service.config["connectTimeout"].(int)
	s.AckTimeout = service.config["ackTimeout"].(int)
	s.TimeoutRetries = service.config["timeoutRetries"].(int)
	return nil
}

func (s *MQTTChannel) Start() error {
	var err error
	if s.ssl {
		cert, err := tls.LoadX509KeyPair(s.service.crt, s.service.key)
		if err != nil {
			return err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cert}}
		log.Printf("Listen MQTT SSL Port%s", s.port)
		s.ln, err = tls.Listen("tcp", s.port, config)
	} else {
		log.Printf("Listen MQTT Port%s", s.port)
		s.ln, err = net.Listen("tcp", s.port)
	}
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := s.ln.Accept()
			if err == nil {
				log.Printf("Accept %s => %s ", conn.RemoteAddr(), s.port)
				go s.AddChannel(conn)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}()
	return nil
}

func (s *MQTTChannel) Stop() {
	s.ln.Close()
}
