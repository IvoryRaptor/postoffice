package source

import (
	"bytes"
	"fmt"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt"
	"github.com/gorilla/websocket"
	"github.com/surge/glog"
	"log"
	"net"
	"net/http"
	"time"
)

type WebSocketChannel struct {
	mqtt.MQTT
	kernel  postoffice.IPostOffice
	server  *http.Server
	ssl     bool
	service *Service
}

type WebSocketConn struct {
	readBuffer bytes.Buffer
	conn       *websocket.Conn
}

func (w *WebSocketConn) Write(b []byte) (n int, err error) {
	err = w.conn.WriteMessage(2, b)
	return len(b), err
}

func (w *WebSocketConn) Read(b []byte) (n int, err error) {
	n, err = w.readBuffer.Read(b)
	if n < len(b) {
		t, p, e := w.conn.ReadMessage()
		//只有二进制可用
		if t != 2 {
			w.Close()
			return -1, nil
		}
		c := copy(b[n:], p)
		if c < len(p) {
			w.readBuffer.Write(p[c:])
		}
		n = n + c
		return n, e
	}
	return n, nil
}

func (w *WebSocketConn) Close() error {
	return w.conn.Close()
}

func (w *WebSocketConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WebSocketConn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *WebSocketConn) SetDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *WebSocketConn) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *WebSocketConn) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}

func (s *WebSocketChannel) Config(service *Service, config map[interface{}]interface{}) error {
	s.service = service
	s.Kernel = service.kernel.(postoffice.IPostOffice)
	s.KeepAlive = service.config["keepAlive"].(int)
	s.ConnectTimeout = service.config["connectTimeout"].(int)
	s.AckTimeout = service.config["ackTimeout"].(int)
	s.TimeoutRetries = service.config["timeoutRetries"].(int)

	var upgrader = websocket.Upgrader{
		Subprotocols: []string{"mqttv3.1", "mqtt", "mqttv3.1.1"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	s.server = &http.Server{Addr: fmt.Sprintf(":%d", config["port"].(int))}
	s.ssl = config["ssl"].(bool)

	http.HandleFunc("/mqtt", func(res http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		glog.Info("new client.........")
		c.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ConnectTimeout)))
		s.AddChannel(&WebSocketConn{conn: c})
	})
	return nil
}

func (s *WebSocketChannel) Start() error {
	go func() {
		if s.ssl {
			crt, key := s.service.crt, s.service.key
			log.Printf("Listen WebSocket SSL Port%s", s.server.Addr)
			if err := s.server.ListenAndServeTLS(crt, key); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err)
			}
		} else {
			log.Printf("Listen WebSocket Port%s", s.server.Addr)
			if err := s.server.ListenAndServe(); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err)
			}
		}
	}()
	return nil
}

func (s *WebSocketChannel) Stop() {
	s.server.Shutdown(nil)
}
