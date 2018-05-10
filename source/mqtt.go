package source

import (
	"github.com/IvoryRaptor/postoffice"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"github.com/IvoryRaptor/postoffice/mqtt"
)

type MQTTSource struct {
	mqtt   mqtt.MQTT
	kernel postoffice.IKernel
	ssl    bool
	config *tls.Config
	port   string
	ln     net.Listener
}

func (s *MQTTSource) Config(kernel postoffice.IKernel, config map[string]interface{}) error {
	s.kernel = kernel
	s.mqtt.Config(kernel, config)
	s.ssl = config["ssl"].(bool)
	s.port = fmt.Sprintf(":%d", config["port"].(int))
	return nil
}

func (s *MQTTSource) Start() error {
	err:= s.mqtt.Start()
	if err != nil {
		return err
	}
	if s.ssl {
		crt, key := s.kernel.GetSSL()
		cert, err := tls.LoadX509KeyPair(crt, key)
		if err != nil {
			return err
		}
		s.config = &tls.Config{Certificates: []tls.Certificate{cert}}
		log.Printf("Listen ssl %s", s.port)
		s.ln, err = tls.Listen("tcp", s.port, s.config)
	} else {
		log.Printf("Listen tcp %s", s.port)
		s.ln, err = net.Listen("tcp", s.port)
	}
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := s.ln.Accept()
			log.Printf("Accept %s => %s ", conn.RemoteAddr(), s.port)
			go s.mqtt.AddChannel(conn)
			if err != nil {
				log.Println(err)
				continue
			}

		}
	}()
	return nil
}

func (s *MQTTSource) Stop(){
	s.mqtt.Stop()
}