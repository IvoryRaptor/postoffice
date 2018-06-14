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
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"github.com/surge/glog"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/postoffice"
)

type (
	OnCompleteFunc func(msg, ack message.Message, err error) error
	OnPublishFunc  func(msg *message.PublishMessage) error
)

type stat struct {
	bytes int64
	msgs  int64
}

func (this *stat) increment(n int64) {
	atomic.AddInt64(&this.bytes, n)
	atomic.AddInt64(&this.msgs, 1)
}

var (
	gsvcid uint64 = 0
)

type Client struct {
	// The number of seconds to keep the connection live if there's no data.
	// If not set then default to 5 mins.
	keepAlive int

	// The number of seconds to wait for the CONNACK message before disconnecting.
	// If not set then default to 2 seconds.
	connectTimeout int

	// The number of seconds to wait for any ACK messages before failing.
	// If not set then default to 20 seconds.
	ackTimeout int

	// The number of times to retry sending a packet if ACK is not received.
	// If no set then default to 3 retries.
	timeoutRetries int

	// Network connection for this Client
	conn io.Closer

	// Wait for the various goroutines to finish starting and stopping
	wgStarted sync.WaitGroup
	wgStopped sync.WaitGroup

	// writeMessage mutex - serializes writes to the outgoing buffer.
	wmu sync.Mutex

	// Whether this is Client is closed or not.
	closed int64

	// Quit signal for determining when this Client should end. If channel is closed,
	// then exit.
	done chan struct{}

	// Incoming data buffer. Bytes are read from the connection and put in here.
	in *buffer

	// Outgoing data buffer. Bytes written here are in turn written out to the connection.
	out *buffer

	// onpub is the method that gets added to the topic subscribers list by the
	// processSubscribe() method. When the server finishes the ack cycle for a
	// PUBLISH message, it will call the subscriber, which is this method.
	//
	// For the server, when this method is called, it means there's a message that
	// should be published to the Client on the other end of this connection. So we
	// will call Publish() to send the message.
	onpub OnPublishFunc

	inStat  stat
	outStat stat

	intmp  []byte
	outtmp []byte

	subs    []interface{}
	qoss    []byte
	rmsgs   []*message.PublishMessage
	kernel  postoffice.IPostOffice
	channel *postoffice.ChannelConfig
}

func (c *Client) GetChannel() *postoffice.ChannelConfig {
	return c.channel
}

func (c *Client) start() error {
	var err error

	// Create the incoming ring buffer
	c.in, err = newBuffer(defaultBufferSize)
	if err != nil {
		return err
	}

	// Create the outgoing ring buffer
	c.out, err = newBuffer(defaultBufferSize)
	if err != nil {
		return err
	}

	// If c is a server
	c.onpub = func(msg *message.PublishMessage) error {
		if err := c.Publish(msg); err != nil {
			glog.Errorf("Client/onPublish: Error publishing message: %v", err)
			return err
		}

		return nil
	}
	// Processor is responsible for reading messages out of the buffer and processing
	// them accordingly.
	c.wgStarted.Add(1)
	c.wgStopped.Add(1)
	go c.processor()

	// Receiver is responsible for reading from the connection and putting data into
	// a buffer.
	c.wgStarted.Add(1)
	c.wgStopped.Add(1)
	go c.receiver()

	// Sender is responsible for writing data in the buffer into the connection.
	c.wgStarted.Add(1)
	c.wgStopped.Add(1)
	go c.sender()

	// Wait for all the goroutines to start before returning
	c.wgStarted.Wait()

	return nil
}

// FIXME: The order of closing here causes panic sometimes. For example, if receiver
// calls this, and closes the buffers, somehow it causes buffer.go:476 to panid.
func (c *Client) Stop() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", c.cid(), r)
		}
	}()

	doit := atomic.CompareAndSwapInt64(&c.closed, 0, 1)
	if !doit {
		return
	}

	// Close quit channel, effectively telling all the goroutines it's time to quit
	if c.done != nil {
		glog.Debugf("(%s) closing c.done", c.cid())
		close(c.done)
	}

	// Close the network connection
	if c.conn != nil {
		glog.Debugf("(%s) closing c.conn", c.cid())
		c.conn.Close()
	}

	c.in.Close()
	c.out.Close()

	// Wait for all the goroutines to stop.
	c.wgStopped.Wait()

	glog.Debugf("(%s) Received %d bytes in %d messages.", c.cid(), c.inStat.bytes, c.inStat.msgs)
	glog.Debugf("(%s) Sent %d bytes in %d messages.", c.cid(), c.outStat.bytes, c.outStat.msgs)

	// Unsubscribe from all the topics for c Client, only for the server side though
	//if !c.Client && c.sess != nil {
	//	topics, _, err := c.sess.Topics()
	//	if err != nil {
	//		glog.Errorf("(%s/%d): %v", c.cid(), c.id, err)
	//	} else {
	//		for _, t := range topics {
	//			if err := c.topicsMgr.Unsubscribe([]byte(t), &c.onpub); err != nil {
	//				glog.Errorf("(%s): Error unsubscribing topic %q: %v", c.cid(), t, err)
	//			}
	//		}
	//	}
	//}
	//
	//// Publish will message if WillFlag is set. Server side only.
	//if !c.Client && c.sess.Cmsg.WillFlag() {
	//	glog.Infof("(%s) Client/stop: connection unexpectedly closed. Sending Will.", c.cid())
	//	c.onPublish(c.sess.Will)
	//}
	//
	//// Remove the Client topics manager
	//if c.Client {
	//	topics.Unregister(c.sess.ID())
	//}
	//
	//// Remove the session from session store if it's suppose to be clean session
	//if c.sess.Cmsg.CleanSession() && c.sessMgr != nil {
	//	c.sessMgr.Del(c.sess.ID())
	//}

	c.conn = nil
	c.in = nil
	c.out = nil
}

func (c *Client) Publish(msg *message.PublishMessage) error {
	//glog.Debugf("Client/Publish: Publishing %s", msg)
	_, err := c.writeMessage(msg)
	if err != nil {
		return fmt.Errorf("(%s) Error sending %s message: %v", c.cid(), msg.Name(), err)
	}

	switch msg.QoS() {
	case message.QosAtMostOnce:
		//if onComplete != nil {
		//	return onComplete(msg, nil, nil)
		//}

		return nil

	case message.QosAtLeastOnce:
		//return c.sess.Pub1ack.Wait(msg, onComplete)

	case message.QosExactlyOnce:
		//return c.sess.Pub2out.Wait(msg, onComplete)
	}

	return nil
}

func (c *Client) subscribe(msg *message.SubscribeMessage, onComplete OnCompleteFunc, onPublish OnPublishFunc) error {
	if onPublish == nil {
		return fmt.Errorf("onPublish function is nil. No need to subscribe.")
	}

	_, err := c.writeMessage(msg)
	if err != nil {
		return fmt.Errorf("(%s) Error sending %s message: %v", c.cid(), msg.Name(), err)
	}

	var _ OnCompleteFunc = func(msg, ack message.Message, err error) error {
		onComplete := onComplete
		//onPublish := onPublish
		if err != nil {
			if onComplete != nil {
				return onComplete(msg, ack, err)
			}
			return err
		}

		sub, ok := msg.(*message.SubscribeMessage)
		if !ok {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Invalid SubscribeMessage received"))
			}
			return nil
		}

		suback, ok := ack.(*message.SubackMessage)
		if !ok {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Invalid SubackMessage received"))
			}
			return nil
		}

		if sub.PacketId() != suback.PacketId() {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Sub and Suback packet ID not the same. %d != %d.", sub.PacketId(), suback.PacketId()))
			}
			return nil
		}

		retcodes := suback.ReturnCodes()
		topics := sub.Topics()

		if len(topics) != len(retcodes) {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Incorrect number of return codes received. Expecting %d, got %d.", len(topics), len(retcodes)))
			}
			return nil
		}

		var err2 error = nil

		for i, t := range topics {
			c := retcodes[i]

			if c == message.QosFailure {
				err2 = fmt.Errorf("Failed to subscribe to '%s'\n%v", string(t), err2)
			} else {
				//c.sess.AddTopic(string(t), c)
				//_, err := c.topicsMgr.Subscribe(t, c, &onPublish)
				//if err != nil {
				//	err2 = fmt.Errorf("Failed to subscribe to '%s' (%v)\n%v", string(t), err, err2)
				//}
			}
		}

		if onComplete != nil {
			return onComplete(msg, ack, err2)
		}

		return err2
	}
	return nil
	//return c.sess.Suback.Wait(msg, onc)
}

func (c *Client) unsubscribe(msg *message.UnsubscribeMessage, onComplete OnCompleteFunc) error {
	_, err := c.writeMessage(msg)
	if err != nil {
		return fmt.Errorf("(%s) Error sending %s message: %v", c.cid(), msg.Name(), err)
	}
	var _ OnCompleteFunc = func(msg, ack message.Message, err error) error {
		onComplete := onComplete

		if err != nil {
			if onComplete != nil {
				return onComplete(msg, ack, err)
			}
			return err
		}

		unsub, ok := msg.(*message.UnsubscribeMessage)
		if !ok {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Invalid UnsubscribeMessage received"))
			}
			return nil
		}

		unsuback, ok := ack.(*message.UnsubackMessage)
		if !ok {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Invalid UnsubackMessage received"))
			}
			return nil
		}

		if unsub.PacketId() != unsuback.PacketId() {
			if onComplete != nil {
				return onComplete(msg, ack, fmt.Errorf("Unsub and Unsuback packet ID not the same. %d != %d.", unsub.PacketId(), unsuback.PacketId()))
			}
			return nil
		}

		var err2 error = nil

		//for _, tb := range unsub.Topics() {
		//	// Remove all subscribers, which basically it's just c Client, since
		//	// each Client has it's own topic tree.
		//	err := c.topicsMgr.Unsubscribe(tb, nil)
		//	if err != nil {
		//		err2 = fmt.Errorf("%v\n%v", err2, err)
		//	}
		//
		//	c.sess.RemoveTopic(string(tb))
		//}

		if onComplete != nil {
			return onComplete(msg, ack, err2)
		}

		return err2
	}
	return nil
	//return c.sess.Unsuback.Wait(msg, onc)
}

func (c *Client) ping(onComplete OnCompleteFunc) error {
	msg := message.NewPingreqMessage()

	_, err := c.writeMessage(msg)
	if err != nil {
		return fmt.Errorf("(%s) Error sending %s message: %v", c.cid(), msg.Name(), err)
	}
	return nil
	//return c.sess.Pingack.Wait(msg, onComplete)
}

func (c *Client) isDone() bool {
	select {
	case <-c.done:
		return true

	default:
	}

	return false
}

func (c *Client) cid() string {
	return fmt.Sprintf("%s/%s", c.channel.Matrix, c.channel.DeviceName)
}
