// Code generated by protoc-gen-go. DO NOT EDIT.
// source: channel.proto

package protoed

/*
defines messages to represent the messaging channel.
*/

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// represents a messaging channel.
type Channel struct {
	Descriptor_          string    `protobuf:"bytes,1,opt,name=descriptor" json:"descriptor,omitempty"`
	Token                string    `protobuf:"bytes,2,opt,name=token" json:"token,omitempty"`
	Sender               *Entity   `protobuf:"bytes,3,opt,name=sender" json:"sender,omitempty"`
	Recipients           []*Entity `protobuf:"bytes,4,rep,name=recipients" json:"recipients,omitempty"`
	Cc                   []*Entity `protobuf:"bytes,5,rep,name=cc" json:"cc,omitempty"`
	Bcc                  []*Entity `protobuf:"bytes,6,rep,name=bcc" json:"bcc,omitempty"`
	Domain               string    `protobuf:"bytes,7,opt,name=domain" json:"domain,omitempty"`
	MinPeriod            float32   `protobuf:"fixed32,8,opt,name=min_period,json=minPeriod" json:"min_period,omitempty"`
	MaxSize              int32     `protobuf:"varint,9,opt,name=max_size,json=maxSize" json:"max_size,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *Channel) Reset()         { *m = Channel{} }
func (m *Channel) String() string { return proto.CompactTextString(m) }
func (*Channel) ProtoMessage()    {}
func (*Channel) Descriptor() ([]byte, []int) {
	return fileDescriptor_channel_3a2c204eaf1a3b56, []int{0}
}
func (m *Channel) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Channel.Unmarshal(m, b)
}
func (m *Channel) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Channel.Marshal(b, m, deterministic)
}
func (dst *Channel) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Channel.Merge(dst, src)
}
func (m *Channel) XXX_Size() int {
	return xxx_messageInfo_Channel.Size(m)
}
func (m *Channel) XXX_DiscardUnknown() {
	xxx_messageInfo_Channel.DiscardUnknown(m)
}

var xxx_messageInfo_Channel proto.InternalMessageInfo

func (m *Channel) GetDescriptor_() string {
	if m != nil {
		return m.Descriptor_
	}
	return ""
}

func (m *Channel) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *Channel) GetSender() *Entity {
	if m != nil {
		return m.Sender
	}
	return nil
}

func (m *Channel) GetRecipients() []*Entity {
	if m != nil {
		return m.Recipients
	}
	return nil
}

func (m *Channel) GetCc() []*Entity {
	if m != nil {
		return m.Cc
	}
	return nil
}

func (m *Channel) GetBcc() []*Entity {
	if m != nil {
		return m.Bcc
	}
	return nil
}

func (m *Channel) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func (m *Channel) GetMinPeriod() float32 {
	if m != nil {
		return m.MinPeriod
	}
	return 0
}

func (m *Channel) GetMaxSize() int32 {
	if m != nil {
		return m.MaxSize
	}
	return 0
}

// represents a sender or recipient of an email.
type Entity struct {
	Email                string   `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Entity) Reset()         { *m = Entity{} }
func (m *Entity) String() string { return proto.CompactTextString(m) }
func (*Entity) ProtoMessage()    {}
func (*Entity) Descriptor() ([]byte, []int) {
	return fileDescriptor_channel_3a2c204eaf1a3b56, []int{1}
}
func (m *Entity) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Entity.Unmarshal(m, b)
}
func (m *Entity) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Entity.Marshal(b, m, deterministic)
}
func (dst *Entity) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Entity.Merge(dst, src)
}
func (m *Entity) XXX_Size() int {
	return xxx_messageInfo_Entity.Size(m)
}
func (m *Entity) XXX_DiscardUnknown() {
	xxx_messageInfo_Entity.DiscardUnknown(m)
}

var xxx_messageInfo_Entity proto.InternalMessageInfo

func (m *Entity) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *Entity) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func init() {
	proto.RegisterType((*Channel)(nil), "protoed.channel.Channel")
	proto.RegisterType((*Entity)(nil), "protoed.channel.Entity")
}

func init() { proto.RegisterFile("channel.proto", fileDescriptor_channel_3a2c204eaf1a3b56) }

var fileDescriptor_channel_3a2c204eaf1a3b56 = []byte{
	// 263 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0x49, 0xda, 0x24, 0xcd, 0x88, 0x08, 0x83, 0xe8, 0x7a, 0x50, 0x42, 0x2f, 0xc6, 0x4b,
	0x84, 0x7a, 0xf0, 0x01, 0xc4, 0xbb, 0xc4, 0x07, 0x28, 0xdb, 0xcd, 0x80, 0x83, 0xdd, 0xd9, 0xb0,
	0xd9, 0x43, 0xed, 0x4b, 0xfa, 0x4a, 0xd2, 0x4d, 0x0e, 0xc5, 0x43, 0x4e, 0x3b, 0xff, 0xbf, 0xdf,
	0xf0, 0xef, 0xfe, 0x70, 0x69, 0xbe, 0xb4, 0x08, 0xed, 0x9b, 0xde, 0xbb, 0xe0, 0xf0, 0x2a, 0x1e,
	0xd4, 0x35, 0x93, 0xbd, 0xfe, 0x4d, 0xa1, 0x78, 0x1b, 0x67, 0x7c, 0x00, 0xe8, 0x68, 0x30, 0x9e,
	0xfb, 0xe0, 0xbc, 0x4a, 0xaa, 0xa4, 0x2e, 0xdb, 0x33, 0x07, 0xaf, 0x21, 0x0b, 0xee, 0x9b, 0x44,
	0xa5, 0xf1, 0x6a, 0x14, 0xf8, 0x0c, 0xf9, 0x40, 0xd2, 0x91, 0x57, 0x8b, 0x2a, 0xa9, 0x2f, 0x36,
	0xb7, 0xcd, 0xbf, 0x8c, 0xe6, 0x5d, 0x02, 0x87, 0x9f, 0x76, 0xc2, 0xf0, 0x15, 0xc0, 0x93, 0xe1,
	0x9e, 0x49, 0xc2, 0xa0, 0x96, 0xd5, 0x62, 0x6e, 0xe9, 0x0c, 0xc5, 0x47, 0x48, 0x8d, 0x51, 0xd9,
	0xfc, 0x42, 0x6a, 0x0c, 0x3e, 0xc1, 0x62, 0x67, 0x8c, 0xca, 0xe7, 0xc9, 0x13, 0x83, 0x37, 0x90,
	0x77, 0xce, 0x6a, 0x16, 0x55, 0xc4, 0x4f, 0x4d, 0x0a, 0xef, 0x01, 0x2c, 0xcb, 0xb6, 0x27, 0xcf,
	0xae, 0x53, 0xab, 0x2a, 0xa9, 0xd3, 0xb6, 0xb4, 0x2c, 0x1f, 0xd1, 0xc0, 0x3b, 0x58, 0x59, 0x7d,
	0xd8, 0x0e, 0x7c, 0x24, 0x55, 0x56, 0x49, 0x9d, 0xb5, 0x85, 0xd5, 0x87, 0x4f, 0x3e, 0xd2, 0x7a,
	0x03, 0xf9, 0x18, 0x70, 0xea, 0x8b, 0xac, 0xe6, 0xfd, 0x54, 0xe5, 0x28, 0x10, 0x61, 0x29, 0xda,
	0xd2, 0x54, 0x62, 0x9c, 0x77, 0x79, 0x7c, 0xe2, 0xcb, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfb,
	0x0f, 0x00, 0x6e, 0xae, 0x01, 0x00, 0x00,
}
