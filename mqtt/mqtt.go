package mqtt

import (
	"errors"
	"fmt"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"log"
	"net"
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
	Kernel         postoffice.IPostOffice
	ConnectTimeout int
	AckTimeout     int
	TimeoutRetries int
	KeepAlive      int
}

func (m *MQTT) AddChannel(conn net.Conn) (err error) {
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
	println(req.String())
	channel := m.Kernel.Authenticate(req)
	if channel == nil {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		err = errors.New("Authenticate error")
		return errors.New("Authenticate error")
		//return err
	}

	m.Kernel.Publish(channel, "device", "online", []byte(channel.Token))
	fmt.Printf("%s device online\n", string([]byte(channel.DeviceName)))
	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	svc := &Client{
		keepAlive:      int(req.KeepAlive()),
		connectTimeout: m.ConnectTimeout,
		ackTimeout:     m.AckTimeout,
		timeoutRetries: m.TimeoutRetries,
		conn:           conn,
		kernel:         m.Kernel,
		channel:        channel,
	}
	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(conn, resp); err != nil {
		return err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	m.Kernel.AddDevice(channel.Matrix+"/"+channel.DeviceName, svc)
	if err := svc.start(); err != nil {
		svc.Stop()
		return err
	}
	return nil
}
