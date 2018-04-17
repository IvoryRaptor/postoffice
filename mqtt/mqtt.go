package mqtt

import (
	"net"
	"time"
	"errors"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"log"
	"sync/atomic"
)

var (
	ErrInvalidConnectionType  error = errors.New("client: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("client: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("client: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("client: buffer has insufficient data.")
)

const (
	minKeepAlive = 30
)

type MQTT struct {
	kernel postoffice.IKernel
	config * Config
}

func (m * MQTT)AddChannel(conn net.Conn) (err error){
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(m.config.ConnectTimeout)))
	resp := message.NewConnackMessage()
	req, err := getConnectMessage(conn)
	if err != nil {
		log.Printf("request connection %s",err.Error())
		if cerr, ok := err.(message.ConnackCode); ok {
			log.Printf("request connection %s",err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return err
	}

	// Authenticate the user, if error, return error and exit
	if err = m.kernel.Authenticate(req); err != nil {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		return err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	svc := &client{
		id:     atomic.AddUint64(&gsvcid, 1),
		client: false,

		keepAlive:      int(req.KeepAlive()),
		connectTimeout: m.config.ConnectTimeout,
		ackTimeout:     m.config.AckTimeout,
		timeoutRetries: m.config.TimeoutRetries,
		actor:          string(req.ClientId()),
		conn:           conn,
		kernel: m.kernel,
		//sessMgr:   this.sessMgr,
		//topicsMgr: this.topicsMgr,
	}
	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(conn, resp); err != nil {
		return err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	if err := svc.start(); err != nil {
		svc.stop()
		return err
	}
	//
	//svc.inStat.increment(int64(req.Len()))
	//svc.outStat.increment(int64(resp.Len()))
	//
	//if err := svc.start(); err != nil {
	//	svc.stop()
	//	return nil, err
	//}

	//this.mu.Lock()
	//this.svcs = append(this.svcs, svc)
	//this.mu.Unlock()
	return nil
}
