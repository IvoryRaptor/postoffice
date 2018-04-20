package source

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"github.com/IvoryRaptor/postoffice"
	"fmt"
	"time"
	"net"
	"bytes"
)

type WebSocketSource struct {
	kernel postoffice.IKernel
	server * http.Server
	ssl bool
	crt string
	key string
}

type WebSocketConn struct {
	readBuffer bytes.Buffer
	readCount  int
	conn       *websocket.Conn
}


func (w *WebSocketConn)Write(b []byte) (n int, err error) {
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

func (w *WebSocketConn) Close() error{
	return w.conn.Close()
}

func (w *WebSocketConn) LocalAddr() net.Addr{
	return w.conn.LocalAddr()
}

func (w *WebSocketConn) RemoteAddr() net.Addr{
	return w.conn.RemoteAddr()
}

func (w *WebSocketConn) SetDeadline(t time.Time) error{
	return w.conn.SetReadDeadline(t)
}

func (w *WebSocketConn) SetReadDeadline(t time.Time) error{
	return w.conn.SetReadDeadline(t)
}

func (w *WebSocketConn) SetWriteDeadline(t time.Time) error{
	return w.conn.SetWriteDeadline(t)
}

func (w *WebSocketSource) Config(kernel postoffice.IKernel, config Config,crt string,key string) error {
	w.key = key
	w.crt = crt
	w.kernel = kernel
	var upgrader = websocket.Upgrader{
		Subprotocols: []string{"mqttv3.1", "mqtt_back", "mqttv3.1.1"},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	w.server = &http.Server{Addr: fmt.Sprintf(":%d", config.Port)}
	w.ssl = config.SSL

	http.HandleFunc("/mqtt", func(res http.ResponseWriter, req *http.Request) {

		//chDone := make(chan bool)
		c, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		w.kernel.AddChannel(&WebSocketConn{conn: c})
	})
	return nil
}

func (w *WebSocketSource) Start() error {
	go func() {
		if w.ssl {
			if err := w.server.ListenAndServeTLS(w.crt, w.key); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err)
			}
		} else {
			if err := w.server.ListenAndServe(); err != nil {
				log.Printf("Httpserver: ListenAndServe() error: %s", err)
			}
		}
	}()
	return nil
}

func (w *WebSocketSource) Stop(){
	w.server.Shutdown(nil)
}
