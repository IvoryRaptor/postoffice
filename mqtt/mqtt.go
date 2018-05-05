package mqtt

import (
	"net"
	"errors"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"log"
)

var (
	ErrInvalidConnectionType  error = errors.New("Client: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("Client: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("Client: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("Client: buffer has insufficient data.")
)

const (
	minKeepAlive = 30
)

type MQTT struct {
	kernel postoffice.IKernel
	config map[string]interface{}
}

func (m * MQTT)AddChannel(conn net.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	resp := message.NewConnackMessage()
	req, err := getConnectMessage(conn)
	if err != nil {
		log.Printf("request connection %s", err.Error())
		if cerr, ok := err.(message.ConnackCode); ok {
			log.Printf("request connection %s", err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return err
	}
	// Authenticate the user, if error, return error and exit
	channel := m.kernel.Authenticate(req)
	if channel == nil {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		return err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	svc := &Client{
		keepAlive:      int(req.KeepAlive()),
		connectTimeout: m.config["connectTimeout"].(int),
		ackTimeout:     m.config["ackTimeout"].(int),
		timeoutRetries: m.config["timeoutRetries"].(int),
		conn:           conn,
		kernel:         m.kernel,
		channel:        channel,
	}
	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(conn, resp); err != nil {
		return err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	m.kernel.AddClient(channel.ClientId, svc)
	if err := svc.start(); err != nil {
		svc.stop()
		return err
	}
	return nil
}
