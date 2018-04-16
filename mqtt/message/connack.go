package message

import "fmt"

// The CONNACK Packet is the packet sent by the Server in response to a CONNECT Packet
// received from a Client. The first packet sent from the Server to the Client MUST
// be a CONNACK Packet [MQTT-3.2.0-1].
//
// If the Client does not receive a CONNACK Packet from the Server within a reasonable
// amount of time, the Client SHOULD close the Network Connection. A "reasonable" amount
// of time depends on the type of application and the communications infrastructure.
type ConnackMessage struct {
	header
	sessionPresent bool
	returnCode     ConnackCode
}

var _ Message = (*ConnackMessage)(nil)

// NewConnackMessage creates a new CONNACK message
func NewConnackMessage() *ConnackMessage {
	msg := &ConnackMessage{}
	msg.SetType(CONNACK)
	return msg
}

// String returns a string representation of the CONNACK message
func (m ConnackMessage) String() string {
	return fmt.Sprintf("%s, Session Present=%t, Return code=%q\n", m.header, m.sessionPresent, m.returnCode)
}

// SessionPresent returns the session present flag value
func (m *ConnackMessage) SessionPresent() bool {
	return m.sessionPresent
}

// SetSessionPresent sets the value of the session present flag
func (m *ConnackMessage) SetSessionPresent(v bool) {
	if v {
		m.sessionPresent = true
	} else {
		m.sessionPresent = false
	}
	m.dirty = true
}

// ReturnCode returns the return code received for the CONNECT message. The return
// type is an error
func (m *ConnackMessage) ReturnCode() ConnackCode {
	return m.returnCode
}

func (m *ConnackMessage) SetReturnCode(ret ConnackCode) {
	m.returnCode = ret
	m.dirty = true
}

func (m *ConnackMessage) Len() int {
	if !m.dirty {
		return len(m.dbuf)
	}

	ml := m.msglen()

	if err := m.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return m.header.msglen() + ml
}

func (m *ConnackMessage) Decode(src []byte) (int, error) {
	total := 0

	n, err := m.header.decode(src)
	total += n
	if err != nil {
		return total, err
	}

	b := src[total]

	if b&254 != 0 {
		return 0, fmt.Errorf("connack/Decode: Bits 7-1 in Connack Acknowledge Flags byte (1) are not 0")
	}

	m.sessionPresent = b&0x1 == 1
	total++

	b = src[total]

	// Read return code
	if b > 5 {
		return 0, fmt.Errorf("connack/Decode: Invalid CONNACK return code (%d)", b)
	}

	m.returnCode = ConnackCode(b)
	total++

	m.dirty = false

	return total, nil
}

func (m *ConnackMessage) Encode(dst []byte) (int, error) {
	if !m.dirty {
		if len(dst) < len(m.dbuf) {
			return 0, fmt.Errorf("connack/Encode: Insufficient buffer size. Expecting %d, got %d.", len(m.dbuf), len(dst))
		}

		return copy(dst, m.dbuf), nil
	}

	// CONNACK remaining length fixed at 2 bytes
	hl := m.header.msglen()
	ml := m.msglen()

	if len(dst) < hl+ml {
		return 0, fmt.Errorf("connack/Encode: Insufficient buffer size. Expecting %d, got %d.", hl+ml, len(dst))
	}

	if err := m.SetRemainingLength(int32(ml)); err != nil {
		return 0, err
	}

	total := 0

	n, err := m.header.encode(dst[total:])
	total += n
	if err != nil {
		return 0, err
	}

	if m.sessionPresent {
		dst[total] = 1
	}
	total++

	if m.returnCode > 5 {
		return total, fmt.Errorf("connack/Encode: Invalid CONNACK return code (%d)", m.returnCode)
	}

	dst[total] = m.returnCode.Value()
	total++

	return total, nil
}

func (m *ConnackMessage) msglen() int {
	return 2
}
