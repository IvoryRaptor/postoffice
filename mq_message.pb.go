// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mq_message.proto

/*
Package postoffice is a generated protocol buffer package.

It is generated from these files:
	mq_message.proto

It has these top-level messages:
	Address
	MQMessage
*/
package postoffice

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Address struct {
	Matrix string `protobuf:"bytes,1,opt,name=matrix" json:"matrix,omitempty"`
	Device string `protobuf:"bytes,2,opt,name=device" json:"device,omitempty"`
}

func (m *Address) Reset()                    { *m = Address{} }
func (m *Address) String() string            { return proto.CompactTextString(m) }
func (*Address) ProtoMessage()               {}
func (*Address) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Address) GetMatrix() string {
	if m != nil {
		return m.Matrix
	}
	return ""
}

func (m *Address) GetDevice() string {
	if m != nil {
		return m.Device
	}
	return ""
}

type MQMessage struct {
	Source      *Address                   `protobuf:"bytes,1,opt,name=source" json:"source,omitempty"`
	Destination *Address                   `protobuf:"bytes,2,opt,name=destination" json:"destination,omitempty"`
	Resource    string                     `protobuf:"bytes,3,opt,name=resource" json:"resource,omitempty"`
	Action      string                     `protobuf:"bytes,4,opt,name=action" json:"action,omitempty"`
	Payload     []byte                     `protobuf:"bytes,5,opt,name=payload,proto3" json:"payload,omitempty"`
	Time        *google_protobuf.Timestamp `protobuf:"bytes,6,opt,name=time" json:"time,omitempty"`
}

func (m *MQMessage) Reset()                    { *m = MQMessage{} }
func (m *MQMessage) String() string            { return proto.CompactTextString(m) }
func (*MQMessage) ProtoMessage()               {}
func (*MQMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *MQMessage) GetSource() *Address {
	if m != nil {
		return m.Source
	}
	return nil
}

func (m *MQMessage) GetDestination() *Address {
	if m != nil {
		return m.Destination
	}
	return nil
}

func (m *MQMessage) GetResource() string {
	if m != nil {
		return m.Resource
	}
	return ""
}

func (m *MQMessage) GetAction() string {
	if m != nil {
		return m.Action
	}
	return ""
}

func (m *MQMessage) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *MQMessage) GetTime() *google_protobuf.Timestamp {
	if m != nil {
		return m.Time
	}
	return nil
}

func init() {
	proto.RegisterType((*Address)(nil), "postoffice.Address")
	proto.RegisterType((*MQMessage)(nil), "postoffice.MQMessage")
}

func init() { proto.RegisterFile("mq_message.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 247 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0x41, 0x4b, 0x03, 0x31,
	0x10, 0x85, 0x59, 0xad, 0x5b, 0x3b, 0xf5, 0x20, 0x11, 0x24, 0xec, 0xc5, 0xd2, 0x53, 0x41, 0x48,
	0x41, 0xf1, 0xe0, 0xd1, 0x1f, 0xd0, 0x83, 0x8b, 0x77, 0x49, 0x37, 0xb3, 0x4b, 0xa0, 0xe9, 0xac,
	0x49, 0x56, 0xf4, 0x47, 0xfb, 0x1f, 0x64, 0x27, 0x59, 0xf5, 0xe2, 0xf1, 0x9b, 0x79, 0x33, 0xef,
	0xf1, 0xe0, 0xd2, 0xbd, 0xbd, 0x3a, 0x0c, 0x41, 0x77, 0xa8, 0x7a, 0x4f, 0x91, 0x04, 0xf4, 0x14,
	0x22, 0xb5, 0xad, 0x6d, 0xb0, 0xba, 0xe9, 0x88, 0xba, 0x03, 0x6e, 0x79, 0xb3, 0x1f, 0xda, 0x6d,
	0xb4, 0x0e, 0x43, 0xd4, 0xae, 0x4f, 0xe2, 0xf5, 0x23, 0xcc, 0x9f, 0x8c, 0xf1, 0x18, 0x82, 0xb8,
	0x86, 0xd2, 0xe9, 0xe8, 0xed, 0x87, 0x2c, 0x56, 0xc5, 0x66, 0x51, 0x67, 0x1a, 0xe7, 0x06, 0xdf,
	0x6d, 0x83, 0xf2, 0x24, 0xcd, 0x13, 0xad, 0xbf, 0x0a, 0x58, 0xec, 0x9e, 0x77, 0xc9, 0x5b, 0xdc,
	0x42, 0x19, 0x68, 0xf0, 0x0d, 0xf2, 0xf5, 0xf2, 0xee, 0x4a, 0xfd, 0xc6, 0x50, 0xd9, 0xa2, 0xce,
	0x12, 0xf1, 0x00, 0x4b, 0x83, 0x21, 0xda, 0xa3, 0x8e, 0x96, 0x8e, 0xfc, 0xf7, 0x9f, 0x8b, 0xbf,
	0x3a, 0x51, 0xc1, 0xb9, 0xc7, 0xec, 0x72, 0xca, 0x59, 0x7e, 0x78, 0x4c, 0xa9, 0x1b, 0xfe, 0x36,
	0x4b, 0x29, 0x13, 0x09, 0x09, 0xf3, 0x5e, 0x7f, 0x1e, 0x48, 0x1b, 0x79, 0xb6, 0x2a, 0x36, 0x17,
	0xf5, 0x84, 0x42, 0xc1, 0x6c, 0x6c, 0x43, 0x96, 0xec, 0x5e, 0xa9, 0x54, 0x95, 0x9a, 0xaa, 0x52,
	0x2f, 0x53, 0x55, 0x35, 0xeb, 0xf6, 0x25, 0x6f, 0xee, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0x43,
	0x8d, 0xdc, 0xfc, 0x72, 0x01, 0x00, 0x00,
}
