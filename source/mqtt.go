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

func (w *MQTTSource) Config(kernel postoffice.IKernel, config map[string]interface{}) error {
	w.kernel = kernel
	w.mqtt.Config(kernel, config)
	w.ssl = config["ssl"].(bool)
	w.port = fmt.Sprintf(":%d", config["port"].(int))
	return nil
}

func (w *MQTTSource) Start() error {
	err:=w.mqtt.Start()
	if err != nil {
		return err
	}
	if w.ssl {
		crt, key := w.kernel.GetSSL()
		cert, err := tls.LoadX509KeyPair(crt, key)
		if err != nil {
			return err
		}
		w.config = &tls.Config{Certificates: []tls.Certificate{cert}}
		log.Printf("Listen ssl %s", w.port)
		w.ln, err = tls.Listen("tcp", w.port, w.config)
	} else {
		log.Printf("Listen tcp %s", w.port)
		w.ln, err = net.Listen("tcp", w.port)
	}
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := w.ln.Accept()
			log.Printf("Accept %s => %s ", conn.RemoteAddr(), w.port)
			w.kernel.AddChannel(conn)
			if err != nil {
				log.Println(err)
				continue
			}

		}
	}()
	return nil
}

func (w *MQTTSource) Stop(){
	w.mqtt.Stop()
}