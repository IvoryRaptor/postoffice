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

package message

import (
	"encoding/binary"
	"fmt"
	"regexp"
)

var clientIdRegexp *regexp.Regexp

func init() {
	// Added space for Paho compliance test
	// Added underscore (_) for MQTT C client test
	clientIdRegexp = regexp.MustCompile("^[0-9a-zA-Z _]*,=/|$")
}

// After a Network Connection is established by a Client to a Server, the first Packet
// sent from the Client to the Server MUST be a CONNECT Packet [MQTT-3.1.0-1].
//
// A Client can only send the CONNECT Packet once over a Network Connection. The Server
// MUST process a second CONNECT Packet sent from a Client as a protocol violation and
// disconnect the Client [MQTT-3.1.0-2].  See section 4.8 for information about
// handling errors.
type ConnectMessage struct {
	header

	// 7: username flag
	// 6: password flag
	// 5: will retain
	// 4-3: will QoS
	// 2: will flag
	// 1: clean session
	// 0: reserved
	connectFlags byte

	version byte

	keepAlive uint16

	protoName,
	clientId,
	willTopic,
	willMessage,
	username,
	password []byte
}

var _ Message = (*ConnectMessage)(nil)

// NewConnectMessage creates a new CONNECT message.
func NewConnectMessage() *ConnectMessage {
	msg := &ConnectMessage{}
	msg.SetType(CONNECT)

	return msg
}

// String returns a string representation of the CONNECT message
func (c ConnectMessage) String() string {
	return fmt.Sprintf("%s, Connect Flags=%08b, Version=%d, KeepAlive=%d, Client ID=%q, Will Topic=%q, Will Message=%q, Username=%q, Password=%q",
		c.header,
		c.connectFlags,
		c.Version(),
		c.KeepAlive(),
		c.ClientId(),
		c.WillTopic(),
		c.WillMessage(),
		c.Username(),
		c.Password(),
	)
}

// Version returns the the 8 bit unsigned value that represents the revision level
// of the protocol used by the Client. The value of the Protocol Level field for
// the version 3.1.1 of the protocol is 4 (0x04).
func (c *ConnectMessage) Version() byte {
	return c.version
}

// SetVersion sets the version value of the CONNECT message
func (c *ConnectMessage) SetVersion(v byte) error {
	if _, ok := SupportedVersions[v]; !ok {
		return fmt.Errorf("connect/SetVersion: Invalid version number %d", v)
	}

	c.version = v
	c.dirty = true

	return nil
}

// CleanSession returns the bit that specifies the handling of the Session state.
// The Client and Server can store Session state to enable reliable messaging to
// continue across a sequence of Network Connections. This bit is used to control
// the lifetime of the Session state.
func (c *ConnectMessage) CleanSession() bool {
	return ((c.connectFlags >> 1) & 0x1) == 1
}

// SetCleanSession sets the bit that specifies the handling of the Session state.
func (c *ConnectMessage) SetCleanSession(v bool) {
	if v {
		c.connectFlags |= 0x2 // 00000010
	} else {
		c.connectFlags &= 253 // 11111101
	}

	c.dirty = true
}

// WillFlag returns the bit that specifies whether a Will Message should be stored
// on the server. If the Will Flag is set to 1 this indicates that, if the Connect
// request is accepted, a Will Message MUST be stored on the Server and associated
// with the Network Connection.
func (c *ConnectMessage) WillFlag() bool {
	return ((c.connectFlags >> 2) & 0x1) == 1
}

// SetWillFlag sets the bit that specifies whether a Will Message should be stored
// on the server.
func (c *ConnectMessage) SetWillFlag(v bool) {
	if v {
		c.connectFlags |= 0x4 // 00000100
	} else {
		c.connectFlags &= 251 // 11111011
	}

	c.dirty = true
}

// WillQos returns the two bits that specify the QoS level to be used when publishing
// the Will Message.
func (c *ConnectMessage) WillQos() byte {
	return (c.connectFlags >> 3) & 0x3
}

// SetWillQos sets the two bits that specify the QoS level to be used when publishing
// the Will Message.
func (c *ConnectMessage) SetWillQos(qos byte) error {
	if qos != QosAtMostOnce && qos != QosAtLeastOnce && qos != QosExactlyOnce {
		return fmt.Errorf("connect/SetWillQos: Invalid QoS level %d", qos)
	}

	c.connectFlags = (c.connectFlags & 231) | (qos << 3) // 231 = 11100111
	c.dirty = true

	return nil
}

// WillRetain returns the bit specifies if the Will Message is to be Retained when it
// is published.
func (c *ConnectMessage) WillRetain() bool {
	return ((c.connectFlags >> 5) & 0x1) == 1
}

// SetWillRetain sets the bit specifies if the Will Message is to be Retained when it
// is published.
func (c *ConnectMessage) SetWillRetain(v bool) {
	if v {
		c.connectFlags |= 32 // 00100000
	} else {
		c.connectFlags &= 223 // 11011111
	}

	c.dirty = true
}

// UsernameFlag returns the bit that specifies whether a user name is present in the
// payload.
func (c *ConnectMessage) UsernameFlag() bool {
	return ((c.connectFlags >> 7) & 0x1) == 1
}

// SetUsernameFlag sets the bit that specifies whether a user name is present in the
// payload.
func (c *ConnectMessage) SetUsernameFlag(v bool) {
	if v {
		c.connectFlags |= 128 // 10000000
	} else {
		c.connectFlags &= 127 // 01111111
	}

	c.dirty = true
}

// PasswordFlag returns the bit that specifies whether a password is present in the
// payload.
func (c *ConnectMessage) PasswordFlag() bool {
	return ((c.connectFlags >> 6) & 0x1) == 1
}

// SetPasswordFlag sets the bit that specifies whether a password is present in the
// payload.
func (c *ConnectMessage) SetPasswordFlag(v bool) {
	if v {
		c.connectFlags |= 64 // 01000000
	} else {
		c.connectFlags &= 191 // 10111111
	}

	c.dirty = true
}

// KeepAlive returns a time interval measured in seconds. Expressed as a 16-bit word,
// it is the maximum time interval that is permitted to elapse between the point at
// which the Client finishes transmitting one Control Packet and the point it starts
// sending the next.
func (c *ConnectMessage) KeepAlive() uint16 {
	return c.keepAlive
}

// SetKeepAlive sets the time interval in which the server should keep the connection
// alive.
func (c *ConnectMessage) SetKeepAlive(v uint16) {
	c.keepAlive = v

	c.dirty = true
}

// ClientId returns an ID that identifies the Client to the Server. Each Client
// connecting to the Server has a unique ClientId. The ClientId MUST be used by
// Clients and by Servers to identify state that they hold relating to this MQTT
// Session between the Client and the Server
func (c *ConnectMessage) ClientId() []byte {
	return c.clientId
}

// SetClientId sets an ID that identifies the Client to the Server.
func (c *ConnectMessage) SetClientId(v []byte) error {
	if len(v) > 0 && !c.validClientId(v) {
		return ErrIdentifierRejected
	}

	c.clientId = v
	c.dirty = true

	return nil
}

// WillTopic returns the topic in which the Will Message should be published to.
// If the Will Flag is set to 1, the Will Topic must be in the payload.
func (c *ConnectMessage) WillTopic() []byte {
	return c.willTopic
}

// SetWillTopic sets the topic in which the Will Message should be published to.
func (c *ConnectMessage) SetWillTopic(v []byte) {
	c.willTopic = v

	if len(v) > 0 {
		c.SetWillFlag(true)
	} else if len(c.willMessage) == 0 {
		c.SetWillFlag(false)
	}

	c.dirty = true
}

// WillMessage returns the Will Message that is to be published to the Will Topic.
func (c *ConnectMessage) WillMessage() []byte {
	return c.willMessage
}

// SetWillMessage sets the Will Message that is to be published to the Will Topic.
func (c *ConnectMessage) SetWillMessage(v []byte) {
	c.willMessage = v

	if len(v) > 0 {
		c.SetWillFlag(true)
	} else if len(c.willTopic) == 0 {
		c.SetWillFlag(false)
	}

	c.dirty = true
}

// Username returns the username from the payload. If the User Name Flag is set to 1,
// this must be in the payload. It can be used by the Server for authentication and
// authorization.
func (c *ConnectMessage) Username() []byte {
	return c.username
}

// SetUsername sets the username for authentication.
func (c *ConnectMessage) SetUsername(v []byte) {
	c.username = v

	if len(v) > 0 {
		c.SetUsernameFlag(true)
	} else {
		c.SetUsernameFlag(false)
	}

	c.dirty = true
}

// Password returns the password from the payload. If the Password Flag is set to 1,
// this must be in the payload. It can be used by the Server for authentication and
// authorization.
func (c *ConnectMessage) Password() []byte {
	return c.password
}

// SetPassword sets the username for authentication.
func (c *ConnectMessage) SetPassword(v []byte) {
	c.password = v

	if len(v) > 0 {
		c.SetPasswordFlag(true)
	} else {
		c.SetPasswordFlag(false)
	}

	c.dirty = true
}

func (c *ConnectMessage) Len() int {
	if !c.dirty {
		return len(c.dbuf)
	}

	ml := c.msglen()

	if err := c.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return c.header.msglen() + ml
}

// For the CONNECT message, the error returned could be a ConnackReturnCode, so
// be sure to check that. Otherwise it's a generic error. If a generic error is
// returned, this Message should be considered invalid.
//
// Caller should call ValidConnackError(err) to see if the returned error is
// a Connack error. If so, caller should send the Client back the corresponding
// CONNACK message.
func (c *ConnectMessage) Decode(src []byte) (int, error) {
	total := 0

	n, err := c.header.decode(src[total:])
	if err != nil {
		return total + n, err
	}
	total += n

	if n, err = c.decodeMessage(src[total:]); err != nil {
		return total + n, err
	}
	total += n

	c.dirty = false

	return total, nil
}

func (c *ConnectMessage) Encode(dst []byte) (int, error) {
	if !c.dirty {
		if len(dst) < len(c.dbuf) {
			return 0, fmt.Errorf("connect/Encode: Insufficient buffer size. Expecting %d, got %d.", len(c.dbuf), len(dst))
		}

		return copy(dst, c.dbuf), nil
	}

	if c.Type() != CONNECT {
		return 0, fmt.Errorf("connect/Encode: Invalid message type. Expecting %d, got %d", CONNECT, c.Type())
	}

	_, ok := SupportedVersions[c.version]
	if !ok {
		return 0, ErrInvalidProtocolVersion
	}

	hl := c.header.msglen()
	ml := c.msglen()

	if len(dst) < hl+ml {
		return 0, fmt.Errorf("connect/Encode: Insufficient buffer size. Expecting %d, got %d.", hl+ml, len(dst))
	}

	if err := c.SetRemainingLength(int32(ml)); err != nil {
		return 0, err
	}

	total := 0

	n, err := c.header.encode(dst[total:])
	total += n
	if err != nil {
		return total, err
	}

	n, err = c.encodeMessage(dst[total:])
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (c *ConnectMessage) encodeMessage(dst []byte) (int, error) {
	total := 0

	n, err := writeLPBytes(dst[total:], []byte(SupportedVersions[c.version]))
	total += n
	if err != nil {
		return total, err
	}

	dst[total] = c.version
	total += 1

	dst[total] = c.connectFlags
	total += 1

	binary.BigEndian.PutUint16(dst[total:], c.keepAlive)
	total += 2

	n, err = writeLPBytes(dst[total:], c.clientId)
	total += n
	if err != nil {
		return total, err
	}

	if c.WillFlag() {
		n, err = writeLPBytes(dst[total:], c.willTopic)
		total += n
		if err != nil {
			return total, err
		}

		n, err = writeLPBytes(dst[total:], c.willMessage)
		total += n
		if err != nil {
			return total, err
		}
	}

	// According to the 3.1 spec, it's possible that the usernameFlag is set,
	// but the username string is missing.
	if c.UsernameFlag() && len(c.username) > 0 {
		n, err = writeLPBytes(dst[total:], c.username)
		total += n
		if err != nil {
			return total, err
		}
	}

	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if c.PasswordFlag() && len(c.password) > 0 {
		n, err = writeLPBytes(dst[total:], c.password)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (c *ConnectMessage) decodeMessage(src []byte) (int, error) {
	var err error
	n, total := 0, 0

	c.protoName, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	c.version = src[total]
	total++

	if verstr, ok := SupportedVersions[c.version]; !ok {
		return total, ErrInvalidProtocolVersion
	} else if verstr != string(c.protoName) {
		return total, ErrInvalidProtocolVersion
	}

	c.connectFlags = src[total]
	total++

	if c.connectFlags&0x1 != 0 {
		return total, fmt.Errorf("connect/decodeMessage: Connect Flags reserved bit 0 is not 0")
	}

	if c.WillQos() > QosExactlyOnce {
		return total, fmt.Errorf("connect/decodeMessage: Invalid QoS level (%d) for %s message", c.WillQos(), c.Name())
	}

	if !c.WillFlag() && (c.WillRetain() || c.WillQos() != QosAtMostOnce) {
		return total, fmt.Errorf("connect/decodeMessage: Protocol violation: If the Will Flag (%t) is set to 0 the Will QoS (%d) and Will Retain (%t) fields MUST be set to zero", c.WillFlag(), c.WillQos(), c.WillRetain())
	}

	if c.UsernameFlag() && !c.PasswordFlag() {
		return total, fmt.Errorf("connect/decodeMessage: Username flag is set but Password flag is not set")
	}

	if len(src[total:]) < 2 {
		return 0, fmt.Errorf("connect/decodeMessage: Insufficient buffer size. Expecting %d, got %d.", 2, len(src[total:]))
	}

	c.keepAlive = binary.BigEndian.Uint16(src[total:])
	total += 2

	c.clientId, n, err = readLPBytes(src[total:])
	total += n
	if err != nil {
		return total, err
	}

	// If the Client supplies a zero-byte ClientId, the Client MUST also set CleanSession to 1
	if len(c.clientId) == 0 && !c.CleanSession() {
		return total, ErrIdentifierRejected
	}

	// The ClientId must contain only characters 0-9, a-z, and A-Z
	// We also support ClientId longer than 23 encoded bytes
	// We do not support ClientId outside of the above characters
	if len(c.clientId) > 0 && !c.validClientId(c.clientId) {
		return total, ErrIdentifierRejected
	}

	if c.WillFlag() {
		c.willTopic, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}

		c.willMessage, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if c.UsernameFlag() && len(src[total:]) > 0 {
		c.username, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if c.PasswordFlag() && len(src[total:]) > 0 {
		c.password, n, err = readLPBytes(src[total:])
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, nil
}

func (c *ConnectMessage) msglen() int {
	total := 0

	verstr, ok := SupportedVersions[c.version]
	if !ok {
		return total
	}

	// 2 bytes protocol name length
	// n bytes protocol name
	// 1 byte protocol version
	// 1 byte connect flags
	// 2 bytes keep alive timer
	total += 2 + len(verstr) + 1 + 1 + 2

	// Add the clientID length, 2 is the length prefix
	total += 2 + len(c.clientId)

	// Add the will topic and will message length, and the length prefixes
	if c.WillFlag() {
		total += 2 + len(c.willTopic) + 2 + len(c.willMessage)
	}

	// Add the username length
	// According to the 3.1 spec, it's possible that the usernameFlag is set,
	// but the user name string is missing.
	if c.UsernameFlag() && len(c.username) > 0 {
		total += 2 + len(c.username)
	}

	// Add the password length
	// According to the 3.1 spec, it's possible that the passwordFlag is set,
	// but the password string is missing.
	if c.PasswordFlag() && len(c.password) > 0 {
		total += 2 + len(c.password)
	}

	return total
}

// validClientId checks the client ID, which is a slice of bytes, to see if it's valid.
// Client ID is valid if it meets the requirement from the MQTT spec:
// 		The Server MUST allow ClientIds which are between 1 and 23 UTF-8 encoded bytes in length,
//		and that contain only the characters
//
//		"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func (c *ConnectMessage) validClientId(cid []byte) bool {
	// Fixed https://github.com/surgemq/surgemq/issues/4
	//if len(cid) > 23 {
	//	return false
	//}

	if c.Version() == 0x3 {
		return true
	}

	return clientIdRegexp.Match(cid)
}
