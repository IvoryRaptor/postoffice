// Copyright (c) 2014 The SurgeMQ Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mqtt

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/surge/glog"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/mqtt/sessions"
	"github.com/IvoryRaptor/postoffice/mqtt/topics"
)

var (
	ErrInvalidConnectionType  error = errors.New("mqtt: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("mqtt: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("mqtt: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("mqtt: buffer has insufficient data.")
)

const (
	DefaultKeepAlive        = 300
	DefaultConnectTimeout   = 2
	DefaultAckTimeout       = 20
	DefaultTimeoutRetries   = 3
	DefaultSessionsProvider = "mem"
	DefaultAuthenticator    = "mockSuccess"
	DefaultTopicsProvider   = "mem"
)

// Server is a library implementation of the MQTT server that, as best it can, complies
// with the MQTT 3.1 and 3.1.1 specs.
type Server struct {
	// The number of seconds to keep the connection live if there's no data.
	// If not set then default to 5 mins.
	KeepAlive int

	// The number of seconds to wait for the CONNECT message before disconnecting.
	// If not set then default to 2 seconds.
	ConnectTimeout int

	// The number of seconds to wait for any ACK messages before failing.
	// If not set then default to 20 seconds.
	AckTimeout int

	// The number of times to retry sending a packet if ACK is not received.
	// If no set then default to 3 retries.
	TimeoutRetries int

	// Authenticator is the authenticator used to check username and password sent
	// in the CONNECT message. If not set then default to "mockSuccess".
	Authenticator string

	// SessionsProvider is the session store that keeps all the Session objects.
	// This is the store to check if CleanSession is set to 0 in the CONNECT message.
	// If not set then default to "mem".
	SessionsProvider string

	// TopicsProvider is the topic store that keeps all the subscription topics.
	// If not set then default to "mem".
	TopicsProvider string

	// authMgr is the authentication manager that we are going to use for authenticating
	// incoming connections
	authMgr auth.IAuthenticator

	// sessMgr is the sessions manager for keeping track of the sessions
	sessMgr *sessions.Manager

	// topicsMgr is the topics manager for keeping track of subscriptions
	topicsMgr *topics.Manager

	// The quit channel for the server. If the server detects that this channel
	// is closed, then it's a signal for it to shutdown as well.
	quit chan struct{}

	ln net.Listener

	// A list of services created by the server. We keep track of them so we can
	// gracefully shut them down if they are still alive when the server goes down.
	svcs []*service

	// Mutex for updating svcs
	mu sync.Mutex

	// A indicator on whether this server is running
	running int32

	// A indicator on whether this server has already checked configuration
	configOnce sync.Once

	subs []interface{}
	qoss []byte
}

// ListenAndServe listents to connections on the URI requested, and handles any
// incoming MQTT client sessions. It should not return until Close() is called
// or if there's some critical error that stops the server from running. The URI
// supplied should be of the form "protocol://host:port" that can be parsed by
// url.Parse(). For example, an URI could be "tcp://0.0.0.0:1883".
func (s *Server) ListenAndServe(uri string) error {
	defer atomic.CompareAndSwapInt32(&s.running, 1, 0)

	if !atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		return fmt.Errorf("server/ListenAndServe: Server is already running")
	}

	s.quit = make(chan struct{})

	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	s.ln, err = net.Listen(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	defer s.ln.Close()

	glog.Infof("server/ListenAndServe: server is ready...")

	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := s.ln.Accept()

		if err != nil {
			// http://zhen.org/blog/graceful-shutdown-of-go-net-dot-listeners/
			select {
			case <-s.quit:
				return nil

			default:
			}

			// Borrowed from go1.3.3/src/pkg/net/http/server.go:1699
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				glog.Errorf("server/ListenAndServe: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		go s.handleConnection(conn)
	}
}

// Publish sends a single MQTT PUBLISH message to the server. On completion, the
// supplied OnCompleteFunc is called. For QOS 0 messages, onComplete is called
// immediately after the message is sent to the outgoing buffer. For QOS 1 messages,
// onComplete is called when PUBACK is received. For QOS 2 messages, onComplete is
// called after the PUBCOMP message is received.
func (s *Server) Publish(msg *message.PublishMessage, onComplete OnCompleteFunc) error {
	if err := s.checkConfiguration(); err != nil {
		return err
	}

	if msg.Retain() {
		if err := s.topicsMgr.Retain(msg); err != nil {
			glog.Errorf("Error retaining message: %v", err)
		}
	}

	if err := s.topicsMgr.Subscribers(msg.Topic(), msg.QoS(), &s.subs, &s.qoss); err != nil {
		return err
	}

	msg.SetRetain(false)

	//glog.Debugf("(server) Publishing to topic %q and %d subscribers", string(msg.Topic()), len(s.subs))
	for _, s := range s.subs {
		if s != nil {
			fn, ok := s.(*OnPublishFunc)
			if !ok {
				glog.Errorf("Invalid onPublish Function")
			} else {
				(*fn)(msg)
			}
		}
	}

	return nil
}

// Close terminates the server by shutting down all the client connections and closing
// the listener. It will, as best it can, clean up after itself.
func (s *Server) Close() error {
	// By closing the quit channel, we are telling the server to stop accepting new
	// connection.
	close(s.quit)

	// We then close the net.Listener, which will force Accept() to return if it's
	// blocked waiting for new connections.
	s.ln.Close()

	for _, svc := range s.svcs {
		glog.Infof("Stopping mqtt %d", svc.id)
		svc.stop()
	}

	if s.sessMgr != nil {
		s.sessMgr.Close()
	}

	if s.topicsMgr != nil {
		s.topicsMgr.Close()
	}

	return nil
}
func (s *Server) AddChannel(c net.Conn) (err error){
	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	c.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ConnectTimeout)))

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
	if err = s.authMgr.Authenticate(req); err != nil {
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
		connectTimeout: s.ConnectTimeout,
		ackTimeout:     s.AckTimeout,
		timeoutRetries: s.TimeoutRetries,

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
// HandleConnection is for the broker to handle an incoming connection from a client
func (s *Server) handleConnection(c io.Closer) (svc *service, err error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}

	defer func() {
		if err != nil {
			c.Close()
		}
	}()

	err = s.checkConfiguration()
	if err != nil {
		return nil, err
	}

	conn, ok := c.(net.Conn)
	if !ok {
		return nil, ErrInvalidConnectionType
	}

	// To establish a connection, we must
	// 1. Read and decode the message.ConnectMessage from the wire
	// 2. If no decoding errors, then authenticate using username and password.
	//    Otherwise, write out to the wire message.ConnackMessage with
	//    appropriate error.
	// 3. If authentication is successful, then either create a new session or
	//    retrieve existing session
	// 4. Write out to the wire a successful message.ConnackMessage message

	// Read the CONNECT message from the wire, if error, then check to see if it's
	// a CONNACK error. If it's CONNACK error, send the proper CONNACK error back
	// to client. Exit regardless of error type.

	conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ConnectTimeout)))

	resp := message.NewConnackMessage()

	req, err := getConnectMessage(conn)

	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			//glog.Debugf("request   message: %s\nresponse message: %s\nerror           : %v", mreq, resp, err)
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return nil, err
	}

	// Authenticate the user, if error, return error and exit
	if err = s.authMgr.Authenticate(req); err != nil {
		resp.SetReturnCode(message.ErrBadUsernameOrPassword)
		resp.SetSessionPresent(false)
		writeMessage(conn, resp)
		return nil, err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(minKeepAlive)
	}

	svc = &service{
		id:     atomic.AddUint64(&gsvcid, 1),
		client: false,

		keepAlive:      int(req.KeepAlive()),
		connectTimeout: s.ConnectTimeout,
		ackTimeout:     s.AckTimeout,
		timeoutRetries: s.TimeoutRetries,

		conn:      conn,
		sessMgr:   s.sessMgr,
		topicsMgr: s.topicsMgr,
	}

	err = s.getSession(svc, req, resp)
	if err != nil {
		return nil, err
	}

	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(c, resp); err != nil {
		return nil, err
	}

	svc.inStat.increment(int64(req.Len()))
	svc.outStat.increment(int64(resp.Len()))

	if err := svc.start(); err != nil {
		svc.stop()
		return nil, err
	}

	//s.mu.Lock()
	//s.svcs = append(s.svcs, svc)
	//s.mu.Unlock()

	glog.Infof("(%s) server/handleConnection: Connection established.", svc.cid())

	return svc, nil
}

func (s *Server) checkConfiguration() error {
	var err error

	s.configOnce.Do(func() {
		if s.KeepAlive == 0 {
			s.KeepAlive = DefaultKeepAlive
		}

		if s.ConnectTimeout == 0 {
			s.ConnectTimeout = DefaultConnectTimeout
		}

		if s.AckTimeout == 0 {
			s.AckTimeout = DefaultAckTimeout
		}

		if s.TimeoutRetries == 0 {
			s.TimeoutRetries = DefaultTimeoutRetries
		}

		if s.Authenticator == "" {
			s.Authenticator = "mockSuccess"
		}

		if err != nil {
			return
		}

		if s.SessionsProvider == "" {
			s.SessionsProvider = "mem"
		}

		s.sessMgr, err = sessions.NewManager(s.SessionsProvider)
		if err != nil {
			return
		}

		if s.TopicsProvider == "" {
			s.TopicsProvider = "mem"
		}

		s.topicsMgr, err = topics.NewManager(s.TopicsProvider)

		return
	})

	return err
}

func (s *Server) getSession(svc *service, req *message.ConnectMessage, resp *message.ConnackMessage) error {
	// If CleanSession is set to 0, the server MUST resume communications with the
	// client based on state from the current session, as identified by the client
	// identifier. If there is no session associated with the client identifier the
	// server must create a new session.
	//
	// If CleanSession is set to 1, the client and server must discard any previous
	// session and start a new one. This session lasts as long as the network c
	// onnection. State data associated with s session must not be reused in any
	// subsequent session.

	var err error

	// Check to see if the client supplied an ID, if not, generate one and set
	// clean session.
	if len(req.ClientId()) == 0 {
		req.SetClientId([]byte(fmt.Sprintf("internalclient%d", svc.id)))
		req.SetCleanSession(true)
	}

	cid := string(req.ClientId())

	// If CleanSession is NOT set, check the session store for existing session.
	// If found, return it.
	if !req.CleanSession() {
		if svc.sess, err = s.sessMgr.Get(cid); err == nil {
			resp.SetSessionPresent(true)

			if err := svc.sess.Update(req); err != nil {
				return err
			}
		}
	}

	// If CleanSession, or no existing session found, then create a new one
	if svc.sess == nil {
		if svc.sess, err = s.sessMgr.New(cid); err != nil {
			return err
		}

		resp.SetSessionPresent(false)

		if err := svc.sess.Init(req); err != nil {
			return err
		}
	}

	return nil
}
