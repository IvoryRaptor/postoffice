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
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/postoffice/mq"
	"strings"
)

var (
	errDisconnect = errors.New("Disconnect")
)

// processor() reads messages from the incoming buffer and processes them
func (c *client) processor() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			//glog.Errorf("(%s) Recovering from panic: %v", c.cid(), r)
		}

		c.wgStopped.Done()
		c.stop()

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

func (c *client) processIncoming(msg message.Message) error {
	var err error = nil

	switch msg := msg.(type) {
	case *message.PublishMessage:
		sp:=strings.Split(string(msg.Topic()),"/")
		if len(sp)<5{
			return nil
		}
		if sp[1]!= c.channel.ProductKey || sp[2]!=c.channel.DeviceName{
			return nil
		}
		action:=sp[3] + "." +sp[4]
		mes := mq.MQMessage{
			Host:     c.kernel.GetHost(),
			Actor:    c.channel.ProductKey + c.channel.DeviceName,
			Resource: sp[3],
			Action:   sp[4],
			Payload:  msg.Payload(),
		}
		matrix,ok := c.kernel.GetMatrix(c.channel.ProductKey)
		if ok {
			println(mes.Actor)
			payload, _ := mes.Descriptor()
			topics, ok := matrix.Action.Load(action)
			if ok {
				for _, topic := range topics.([]string) {
					c.kernel.Publish(topic, payload)
				}
			}
		}else{
			println("miss ")
		}
		//println("Publish:" + c.actor + string(msg.Topic()))

		// For PUBLISH message, we should figure out what QoS it is and process accordingly
		// If QoS == 0, we should just take the next step, no ack required
		// If QoS == 1, we should send back PUBACK, then take the next step
		// If QoS == 2, we need to put it in the ack queue, send back PUBREC
		err = c.processPublish(msg)

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
//
//func (this *client) processAcked(ackq *sessions.Ackqueue) {
//	for _, ackmsg := range ackq.Acked() {
//		// Let's get the messages from the saved message byte slices.
//		msg, err := ackmsg.Mtype.New()
//		if err != nil {
//			glog.Errorf("process/processAcked: Unable to creating new %s message: %v", ackmsg.Mtype, err)
//			continue
//		}
//
//		if _, err := msg.Decode(ackmsg.Msgbuf); err != nil {
//			glog.Errorf("process/processAcked: Unable to decode %s message: %v", ackmsg.Mtype, err)
//			continue
//		}
//
//		ack, err := ackmsg.State.New()
//		if err != nil {
//			glog.Errorf("process/processAcked: Unable to creating new %s message: %v", ackmsg.State, err)
//			continue
//		}
//
//		if _, err := ack.Decode(ackmsg.Ackbuf); err != nil {
//			glog.Errorf("process/processAcked: Unable to decode %s message: %v", ackmsg.State, err)
//			continue
//		}
//
//		//glog.Debugf("(%s) Processing acked message: %v", this.cid(), ack)
//
//		// - PUBACK if it's QoS 1 message. This is on the client side.
//		// - PUBREL if it's QoS 2 message. This is on the server side.
//		// - PUBCOMP if it's QoS 2 message. This is on the client side.
//		// - SUBACK if it's a subscribe message. This is on the client side.
//		// - UNSUBACK if it's a unsubscribe message. This is on the client side.
//		switch ackmsg.State {
//		case message.PUBREL:
//			// If ack is PUBREL, that means the QoS 2 message sent by a remote client is
//			// releassed, so let's publish it to other subscribers.
//			if err = this.onPublish(msg.(*message.PublishMessage)); err != nil {
//				glog.Errorf("(%s) Error processing ack'ed %s message: %v", this.cid(), ackmsg.Mtype, err)
//			}
//
//		case message.PUBACK, message.PUBCOMP, message.SUBACK, message.UNSUBACK, message.PINGRESP:
//			glog.Debugf("process/processAcked: %s", ack)
//			// If ack is PUBACK, that means the QoS 1 message sent by this client got
//			// ack'ed. There's nothing to do other than calling onComplete() below.
//
//			// If ack is PUBCOMP, that means the QoS 2 message sent by this client got
//			// ack'ed. There's nothing to do other than calling onComplete() below.
//
//			// If ack is SUBACK, that means the SUBSCRIBE message sent by this client
//			// got ack'ed. There's nothing to do other than calling onComplete() below.
//
//			// If ack is UNSUBACK, that means the SUBSCRIBE message sent by this client
//			// got ack'ed. There's nothing to do other than calling onComplete() below.
//
//			// If ack is PINGRESP, that means the PINGREQ message sent by this client
//			// got ack'ed. There's nothing to do other than calling onComplete() below.
//
//			err = nil
//
//		default:
//			glog.Errorf("(%s) Invalid ack message type %s.", this.cid(), ackmsg.State)
//			continue
//		}
//
//		// Call the registered onComplete function
//		if ackmsg.OnComplete != nil {
//			onComplete, ok := ackmsg.OnComplete.(OnCompleteFunc)
//			if !ok {
//				glog.Errorf("process/processAcked: Error type asserting onComplete function: %v", reflect.TypeOf(ackmsg.OnComplete))
//			} else if onComplete != nil {
//				if err := onComplete(msg, ack, nil); err != nil {
//					glog.Errorf("process/processAcked: Error running onComplete(): %v", err)
//				}
//			}
//		}
//	}
//}

// For PUBLISH message, we should figure out what QoS it is and process accordingly
// If QoS == 0, we should just take the next step, no ack required
// If QoS == 1, we should send back PUBACK, then take the next step
// If QoS == 2, we need to put it in the ack queue, send back PUBREC
func (c *client) processPublish(msg *message.PublishMessage) error {
	switch msg.QoS() {
	case message.QosExactlyOnce:
		//c.sess.Pub2in.Wait(msg, nil)

		resp := message.NewPubrecMessage()
		resp.SetPacketId(msg.PacketId())

		_, err := c.writeMessage(resp)
		return err

	case message.QosAtLeastOnce:
		resp := message.NewPubackMessage()
		resp.SetPacketId(msg.PacketId())

		if _, err := c.writeMessage(resp); err != nil {
			return err
		}

		return c.onPublish(msg)

	case message.QosAtMostOnce:
		return c.onPublish(msg)
	}

	return fmt.Errorf("(%s) invalid message QoS %d.", c.cid(), msg.QoS())
}

// For SUBSCRIBE message, we should add subscriber, then send back SUBACK
func (c *client) processSubscribe(msg *message.SubscribeMessage) error {
	resp := message.NewSubackMessage()
	resp.SetPacketId(msg.PacketId())

	// Subscribe to the different topics
	var retcodes []byte

	topics := msg.Topics()
	qos := msg.Qos()

	c.rmsgs = c.rmsgs[0:0]

	for i, t := range topics {
		//rqos, err := c.topicsMgr.Subscribe(t, qos[i], &c.onpub)
		//if err != nil {
		//	return err
		//}
		//c.sess.AddTopic(string(t), qos[i])
		//
		//retcodes = append(retcodes, rqos)
		//
		//// yeah I am not checking errors here. If there's an error we don't want the
		//// subscription to stop, just let it go.
		//c.topicsMgr.Retained(t, &c.rmsgs)
		println(string(qos), i, string(t))
		//glog.Debugf("(%s) topic = %s, retained count = %d", c.cid(), string(t), len(c.rmsgs))
	}

	if err := resp.AddReturnCodes(retcodes); err != nil {
		return err
	}

	if _, err := c.writeMessage(resp); err != nil {
		return err
	}

	for _, rm := range c.rmsgs {
		if err := c.publish(rm, nil); err != nil {
			//glog.Errorf("client/processSubscribe: Error publishing retained message: %v", err)
			return err
		}
	}

	return nil
}

// For UNSUBSCRIBE message, we should remove the subscriber, and send back UNSUBACK
func (c *client) processUnsubscribe(msg *message.UnsubscribeMessage) error {
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
// the ack cycle. This method will get the list of subscribers based on the publish
// topic, and publishes the message to the list of subscribers.
func (c *client) onPublish(msg *message.PublishMessage) error {
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
