package source

import (
	"github.com/IvoryRaptor/postoffice"
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

type TcpSource struct {
	kernel postoffice.IKernel
	ssl bool
	config *tls.Config
	port string
	ln net.Listener
}

func (w *TcpSource) Config(kernel postoffice.IKernel, config Config,crt string,key string) error {
	w.kernel = kernel
	w.ssl = config.SSL
	if w.ssl {
		crt, err := tls.LoadX509KeyPair(crt, key)
		if err != nil {
			return err
		}
		w.config = &tls.Config{Certificates: []tls.Certificate{crt}}
	}
	w.port = fmt.Sprintf(":%d", config.Port)
	return nil
}

func (w *TcpSource) Start() error {
	var err error
	if w.ssl {
		log.Printf("Listen ssl %s",w.port)
		w.ln, err = tls.Listen("tcp", w.port, w.config)
	} else {
		log.Printf("Listen tcp %s",w.port)
		w.ln, err = net.Listen("tcp", w.port)
	}
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := w.ln.Accept()
			log.Printf("Accept %s => %s ",conn.RemoteAddr(), w.port)
			w.kernel.AddChannel(conn)
			if err != nil {
				log.Println(err)
				continue
			}

		}
	}()
	return nil
}

func (w *TcpSource) Stop(){

}