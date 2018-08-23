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
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"io"
	"strings"
)

var (
	errDisconnect = errors.New("Disconnect")
)

// processor() reads messages from the incoming buffer and processes them
func (c *Client) processor() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			//glog.Errorf("(%s) Recovering from panic: %v", c.cid(), r)
		}

		c.wgStopped.Done()
		c.Stop()

		//glog.Debugf("(%s) Stopping processor", c.cid())
	}()

	//glog.Debugf("(%s) Starting processor", c.cid())

	c.wgStarted.Done()

	for {
		// 1. Find out what message is next and the size of the message
		mtype, total, err := c.peekMessageSize()
		if err != nil {
			//if err != io.EOF {
			//glog.Errorf("(%s) Error peeking next message size: %v", c.cid(), err)
			//}
			return
		}

		msg, n, err := c.peekMessage(mtype, total)
		if err != nil {
			//if err != io.EOF {
			//glog.Errorf("(%s) Error peeking next message: %v", c.cid(), err)
			//}
			return
		}

		//glog.Debugf("(%s) Received: %s", c.cid(), msg)

		c.inStat.increment(int64(n))

		// 5. Process the read message
		err = c.processIncoming(msg)
		if err != nil {
			if err != errDisconnect {
				//glog.Errorf("(%s) Error processing %s: %v", c.cid(), msg.Name(), err)
			} else {
				return
			}
		}

		// 7. We should commit the bytes in the buffer so we can move on
		_, err = c.in.ReadCommit(total)
		if err != nil {
			if err != io.EOF {
				//glog.Errorf("(%s) Error committing %d read bytes: %v", c.cid(), total, err)
			}
			return
		}

		// 7. Check to see if done is closed, if so, exit
		if c.isDone() && c.in.Len() == 0 {
			return
		}

		//if c.inStat.msgs%1000 == 0 {
		//	glog.Debugf("(%s) Going to process message %d", c.cid(), c.inStat.msgs)
		//}
	}
}

func (c *Client) processIncoming(msg message.Message) error {
	var err error = nil

	switch msg := msg.(type) {
	case *message.PublishMessage:
		sp := strings.Split(string(msg.Topic()), "/")
		c.kernel.Publish(c.channel, sp[3], sp[4], msg.Payload())

	case *message.PubackMessage:
		// For PUBACK message, it means QoS 1, we should send to ack queue
		//c.sess.Pub1ack.Ack(msg)
		//c.processAcked(c.sess.Pub1ack)

	case *message.PubrecMessage:
		// For PUBREC message, it means QoS 2, we should send to ack queue, and send back PUBREL
		//if err = c.sess.Pub2out.Ack(msg); err != nil {
		//	break
		//}
		//println("PubrecMessage", c.actor)
		resp := message.NewPubrelMessage()
		resp.SetPacketId(msg.PacketId())
		_, err = c.writeMessage(resp)

	case *message.PubrelMessage:
		// For PUBREL message, it means QoS 2, we should send to ack queue, and send back PUBCOMP
		//if err = c.sess.Pub2in.Ack(msg); err != nil {
		//	break
		//}
		//
		//c.processAcked(c.sess.Pub2in)

		resp := message.NewPubcompMessage()
		resp.SetPacketId(msg.PacketId())
		_, err = c.writeMessage(resp)

	case *message.PubcompMessage:
		// For PUBCOMP message, it means QoS 2, we should send to ack queue
		//if err = c.sess.Pub2out.Ack(msg); err != nil {
		//	break
		//}
		//
		//c.processAcked(c.sess.Pub2out)

	case *message.SubscribeMessage:
		fmt.Printf("**************%s\n", string(msg.Topics()[0]))

		// For SUBSCRIBE message, we should add subscriber, then send back SUBACK
		return c.processSubscribe(msg)

	case *message.SubackMessage:
		// For SUBACK message, we should send to ack queue
		//c.sess.Suback.Ack(msg)
		//c.processAcked(c.sess.Suback)

	case *message.UnsubscribeMessage:
		// For UNSUBSCRIBE message, we should remove subscriber, then send back UNSUBACK
		return c.processUnsubscribe(msg)

	case *message.UnsubackMessage:
		// For UNSUBACK message, we should send to ack queue
		//c.sess.Unsuback.Ack(msg)
		//c.processAcked(c.sess.Unsuback)

	case *message.PingreqMessage:
		// For PINGREQ message, we should send back PINGRESP
		resp := message.NewPingrespMessage()
		_, err = c.writeMessage(resp)

	case *message.PingrespMessage:
		//c.sess.Pingack.Ack(msg)
		//c.processAcked(c.sess.Pingack)

	case *message.DisconnectMessage:
		// For DISCONNECT message, we should quit
		//c.sess.Cmsg.SetWillFlag(false)
		return errDisconnect

	default:
		return fmt.Errorf("(%s) invalid message type %s.", c.cid(), msg.Name())
	}

	if err != nil {
		//glog.Debugf("(%s) Error processing acked message: %v", c.cid(), err)
	}

	return err
}

// For SUBSCRIBE message, we should add subscriber, then send back SUBACK
func (c *Client) processSubscribe(msg *message.SubscribeMessage) error {
	resp := message.NewSubackMessage()
	resp.SetPacketId(msg.PacketId())

	// Subscribe to the different topics
	var retcodes []byte

	//topics := msg.Topics()
	//qos := msg.Qos()

	c.rmsgs = c.rmsgs[0:0]

	//for i, t := range topics {
	//	println(string(qos), i, string(t))
	//}

	if err := resp.AddReturnCodes(retcodes); err != nil {
		return err
	}

	if _, err := c.writeMessage(resp); err != nil {
		return err
	}

	for _, rm := range c.rmsgs {
		if err := c.Publish(rm); err != nil {
			//glog.Errorf("Client/processSubscribe: Error publishing retained message: %v", err)
			return err
		}
	}

	return nil
}

// For UNSUBSCRIBE message, we should remove the subscriber, and send back UNSUBACK
func (c *Client) processUnsubscribe(msg *message.UnsubscribeMessage) error {
	//topics := msg.Topics()

	//for _, t := range topics {
	//	//c.topicsMgr.Unsubscribe(t, &c.onpub)
	//	//c.sess.RemoveTopic(string(t))
	//}

	resp := message.NewUnsubackMessage()
	resp.SetPacketId(msg.PacketId())

	_, err := c.writeMessage(resp)
	return err
}

// onPublish() is called when the server receives a PUBLISH message AND have completed
// the ack cycle. This method will get the list of subscribers based on the Publish
// topic, and publishes the message to the list of subscribers.
func (c *Client) onPublish(msg *message.PublishMessage) error {
	if msg.Retain() {
		//if err := c.topicsMgr.Retain(msg); err != nil {
		//	glog.Errorf("(%s) Error retaining message: %v", c.cid(), err)
		//}
	}

	//err := c.topicsMgr.Subscribers(msg.Topic(), msg.QoS(), &c.subs, &c.qoss)
	//if err != nil {
	//	glog.Errorf("(%s) Error retrieving subscribers list: %v", c.cid(), err)
	//	return err
	//}

	msg.SetRetain(false)

	//glog.Debugf("(%s) Publishing to topic %q and %d subscribers", c.cid(), string(msg.Topic()), len(c.subs))
	for _, s := range c.subs {
		if s != nil {
			fn, ok := s.(*OnPublishFunc)
			if !ok {
				//glog.Errorf("Invalid onPublish Function")
				return fmt.Errorf("Invalid onPublish Function")
			} else {
				(*fn)(msg)
			}
		}
	}

	return nil
}
