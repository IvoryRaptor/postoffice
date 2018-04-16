package kernel

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/IvoryRaptor/postoffice/source"
	"net"
	"time"
	"sync/atomic"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"errors"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/matrix"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice"
)

const (
	minKeepAlive = 30
)

var (
	ErrInvalidConnectionType  error = errors.New("mqtt: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("mqtt: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("mqtt: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("mqtt: buffer has insufficient data.")
)

type Kernel struct {
	host          int
	ConfigFile    string
	run           bool
	source        []source.ISource
	authenticator auth.IAuthenticator
	matrixManger  matrix.Manager
	config        Config
	mq            mq.IMQ
}

func (kernel *Kernel)IsRun() bool {
	return kernel.run
}

func (kernel *Kernel)GetHost() int{
	return kernel.host
}

func (kernel *Kernel) GetMatrix(name string) (*postoffice.Matrix, bool) {
	return kernel.matrixManger.GetMatrix(name)
}

func (kernel *Kernel) AddChannel(c net.Conn) (err error){

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	c.SetReadDeadline(time.Now().Add(time.Second * time.Duration(kernel.config.MQTT.ConnectTimeout)))

	resp := message.NewConnackMessage()

	req, err := getConnectMessage(c)

	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			//glog.Debugf("request   message: %s\nresponse message: %s\nerror           : %v", mreq, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(c, resp)
		}
		return err
	}

	// Authenticate the user, if error, return error and exit
	if err = kernel.authenticator.Authenticate(req); err != nil {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(c, resp)
		return err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	svc := &service{
		id:     atomic.AddUint64(&gsvcid, 1),
		client: false,

		keepAlive:      int(req.KeepAlive()),
		connectTimeout: kernel.config.MQTT.ConnectTimeout,
		ackTimeout:     kernel.config.MQTT.AckTimeout,
		timeoutRetries: kernel.config.MQTT.TimeoutRetries,

		conn:      c,
		//sessMgr:   kernel.sessMgr,
		//topicsMgr: kernel.topicsMgr,
	}

	//err = kernel.getSession(svc, req, resp)
	//if err != nil {
	//	return err
	//}

	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(c, resp); err != nil {
		return err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	if err := svc.start(); err != nil {
		svc.stop()
		return err
	}
	return nil
	//s.mu.Lock()
	//s.svcs = append(s.svcs, svc)
	//s.mu.Unlock()
}

func (kernel *Kernel) WaitStop() {
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	kernel.Stop()
	stopChan <- struct{}{}
	os.Exit(0)
}
