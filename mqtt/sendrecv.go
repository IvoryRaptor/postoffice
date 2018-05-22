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
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/surge/glog"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
)

type netReader interface {
	io.Reader
	SetReadDeadline(t time.Time) error
}

type timeoutReader struct {
	d    time.Duration
	conn netReader
}

func (r timeoutReader) Read(b []byte) (int, error) {
	if err := r.conn.SetReadDeadline(time.Now().Add(r.d)); err != nil {
		return 0, err
	}
	return r.conn.Read(b)
}

// receiver() reads data from the network, and writes the data into the incoming buffer
func (c *Client) receiver() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", c.cid(), r)
		}

		c.wgStopped.Done()

		glog.Debugf("(%s) Stopping receiver", c.cid())
	}()

	glog.Debugf("(%s) Starting receiver", c.cid())

	c.wgStarted.Done()

	switch conn := c.conn.(type) {
	case net.Conn:
		//glog.Debugf("server/handleConnection: Setting read deadline to %d", time.Second*time.Duration(c.keepAlive))
		keepAlive := time.Second * time.Duration(c.keepAlive)
		r := timeoutReader{
			d:    keepAlive + (keepAlive / 2),
			conn: conn,
		}

		for {
			_, err := c.in.ReadFrom(r)

			if err != nil {
				if err != io.EOF {
					c.kernel.Publish(c.channel, "device", "offline", []byte(c.channel.DeviceName+"@"+c.channel.Token))
					c.kernel.Close(c.cid())
					glog.Errorf("(%s) error reading from connection: %v", c.cid(), err)
				}
				return
			}
		}

	//case *websocket.Conn:
	//	glog.Errorf("(%s) Websocket: %v", c.cid(), ErrInvalidConnectionType)

	default:
		glog.Errorf("(%s) %v", c.cid(), ErrInvalidConnectionType)
	}
}

// sender() writes data from the outgoing buffer to the network
func (c *Client) sender() {
	defer func() {
		// Let's recover from panic
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", c.cid(), r)
		}

		c.wgStopped.Done()

		glog.Debugf("(%s) Stopping sender", c.cid())
	}()

	glog.Debugf("(%s) Starting sender", c.cid())

	c.wgStarted.Done()

	switch conn := c.conn.(type) {
	case net.Conn:
		for {
			_, err := c.out.WriteTo(conn)

			if err != nil {
				if err != io.EOF {
					glog.Errorf("(%s) error writing data: %v", c.cid(), err)
				}
				return
			}
		}

	//case *websocket.Conn:
	//	glog.Errorf("(%s) Websocket not supported", c.cid())

	default:
		glog.Errorf("(%s) Invalid connection type", c.cid())
	}
}

// peekMessageSize() reads, but not commits, enough bytes to determine the size of
// the next message and returns the type and size.
func (c *Client) peekMessageSize() (message.MessageType, int, error) {
	var (
		b   []byte
		err error
		cnt int = 2
	)

	if c.in == nil {
		err = ErrBufferNotReady
		return 0, 0, err
	}

	// Let's read enough bytes to get the message header (msg type, remaining length)
	for {
		// If we have read 5 bytes and still not done, then there's a problem.
		if cnt > 5 {
			return 0, 0, fmt.Errorf("sendrecv/peekMessageSize: 4th byte of remaining length has continuation bit set")
		}

		// Peek cnt bytes from the input buffer.
		b, err = c.in.ReadWait(cnt)
		if err != nil {
			return 0, 0, err
		}

		// If not enough bytes are returned, then continue until there's enough.
		if len(b) < cnt {
			continue
		}

		// If we got enough bytes, then check the last byte to see if the continuation
		// bit is set. If so, increment cnt and continue peeking
		if b[cnt-1] >= 0x80 {
			cnt++
		} else {
			break
		}
	}

	// Get the remaining length of the message
	remlen, m := binary.Uvarint(b[1:])

	// Total message length is remlen + 1 (msg type) + m (remlen bytes)
	total := int(remlen) + 1 + m

	mtype := message.MessageType(b[0] >> 4)

	return mtype, total, err
}

// peekMessage() reads a message from the buffer, but the bytes are NOT committed.
// This means the buffer still thinks the bytes are not read yet.
func (c *Client) peekMessage(mtype message.MessageType, total int) (message.Message, int, error) {
	var (
		b    []byte
		err  error
		i, n int
		msg  message.Message
	)

	if c.in == nil {
		return nil, 0, ErrBufferNotReady
	}

	// Peek until we get total bytes
	for i = 0; ; i++ {
		// Peek remlen bytes from the input buffer.
		b, err = c.in.ReadWait(total)
		if err != nil && err != ErrBufferInsufficientData {
			return nil, 0, err
		}

		// If not enough bytes are returned, then continue until there's enough.
		if len(b) >= total {
			break
		}
	}

	msg, err = mtype.New()
	if err != nil {
		return nil, 0, err
	}

	n, err = msg.Decode(b)
	return msg, n, err
}

// readMessage() reads and copies a message from the buffer. The buffer bytes are
// committed as a result of the read.
func (c *Client) readMessage(mtype message.MessageType, total int) (message.Message, int, error) {
	var (
		b   []byte
		err error
		n   int
		msg message.Message
	)

	if c.in == nil {
		err = ErrBufferNotReady
		return nil, 0, err
	}

	if len(c.intmp) < total {
		c.intmp = make([]byte, total)
	}

	// Read until we get total bytes
	l := 0
	for l < total {
		n, err = c.in.Read(c.intmp[l:])
		l += n
		glog.Debugf("read %d bytes, total %d", n, l)
		if err != nil {
			return nil, 0, err
		}
	}

	b = c.intmp[:total]

	msg, err = mtype.New()
	if err != nil {
		return msg, 0, err
	}

	n, err = msg.Decode(b)
	return msg, n, err
}

// writeMessage() writes a message to the outgoing buffer
func (c *Client) writeMessage(msg message.Message) (int, error) {
	var (
		l    int = msg.Len()
		m, n int
		err  error
		buf  []byte
		wrap bool
	)

	if c.out == nil {
		return 0, ErrBufferNotReady
	}

	// This is to serialize writes to the underlying buffer. Multiple goroutines could
	// potentially get here because of calling Publish() or Subscribe() or other
	// functions that will send messages. For example, if a message is received in
	// another connetion, and the message needs to be published to c Client, then
	// the Publish() function is called, and at the same time, another Client could
	// do exactly the same thing.
	//
	// Not an ideal fix though. If possible we should remove mutex and be lockfree.
	// Mainly because when there's a large number of goroutines that want to Publish
	// to c Client, then they will all block. However, c will do for now.
	//
	// FIXME: Try to find a better way than a mutex...if possible.
	c.wmu.Lock()
	defer c.wmu.Unlock()

	buf, wrap, err = c.out.WriteWait(l)
	if err != nil {
		return 0, err
	}

	if wrap {
		if len(c.outtmp) < l {
			c.outtmp = make([]byte, l)
		}

		n, err = msg.Encode(c.outtmp[0:])
		if err != nil {
			return 0, err
		}

		m, err = c.out.Write(c.outtmp[0:n])
		if err != nil {
			return m, err
		}
	} else {
		n, err = msg.Encode(buf[0:])
		if err != nil {
			return 0, err
		}

		m, err = c.out.WriteCommit(n)
		if err != nil {
			return 0, err
		}
	}

	c.outStat.increment(int64(m))

	return m, nil
}
