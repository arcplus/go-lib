package tool

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// for common list req
type TestProto struct {
	Id                   string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name                 string            `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Age                  int64             `protobuf:"varint,3,opt,name=age,proto3" json:"age,omitempty"`
	NextId               string            `protobuf:"bytes,4,opt,name=next_id,json=nextId,proto3" json:"next_id,omitempty"`
	Filter               map[string]string `protobuf:"bytes,5,rep,name=filter,proto3" json:"filter,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *TestProto) Reset()         { *m = TestProto{} }
func (m *TestProto) String() string { return proto.CompactTextString(m) }
func (*TestProto) ProtoMessage()    {}
func (*TestProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_x_ad4cd39ce127d50b, []int{0}
}
func (m *TestProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestProto.Unmarshal(m, b)
}
func (m *TestProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestProto.Marshal(b, m, deterministic)
}
func (dst *TestProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestProto.Merge(dst, src)
}
func (m *TestProto) XXX_Size() int {
	return xxx_messageInfo_TestProto.Size(m)
}
func (m *TestProto) XXX_DiscardUnknown() {
	xxx_messageInfo_TestProto.DiscardUnknown(m)
}

var xxx_messageInfo_TestProto proto.InternalMessageInfo

func (m *TestProto) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *TestProto) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *TestProto) GetAge() int64 {
	if m != nil {
		return m.Age
	}
	return 0
}

func (m *TestProto) GetNextId() string {
	if m != nil {
		return m.NextId
	}
	return ""
}

func (m *TestProto) GetFilter() map[string]string {
	if m != nil {
		return m.Filter
	}
	return nil
}

func init() {
	proto.RegisterType((*TestProto)(nil), "tool.test_proto")
	proto.RegisterMapType((map[string]string)(nil), "tool.test_proto.FilterEntry")
}

func init() { proto.RegisterFile("x.proto", fileDescriptor_x_ad4cd39ce127d50b) }

var fileDescriptor_x_ad4cd39ce127d50b = []byte{
	// 188 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x62, 0xaf, 0xd0, 0x2b, 0x28,
	0xca, 0x2f, 0xc9, 0x17, 0x62, 0x29, 0xc9, 0xcf, 0xcf, 0x51, 0x3a, 0xc3, 0xc8, 0xc5, 0x55, 0x92,
	0x5a, 0x5c, 0x12, 0x0f, 0x11, 0xe4, 0xe3, 0x62, 0xca, 0x4c, 0x91, 0x60, 0x54, 0x60, 0xd4, 0xe0,
	0x0c, 0x62, 0xca, 0x4c, 0x11, 0x12, 0xe2, 0x62, 0xc9, 0x4b, 0xcc, 0x4d, 0x95, 0x60, 0x02, 0x8b,
	0x80, 0xd9, 0x42, 0x02, 0x5c, 0xcc, 0x89, 0xe9, 0xa9, 0x12, 0xcc, 0x0a, 0x8c, 0x1a, 0xcc, 0x41,
	0x20, 0xa6, 0x90, 0x38, 0x17, 0x7b, 0x5e, 0x6a, 0x45, 0x49, 0x7c, 0x66, 0x8a, 0x04, 0x0b, 0x58,
	0x21, 0x1b, 0x88, 0xeb, 0x99, 0x22, 0x64, 0xc2, 0xc5, 0x96, 0x96, 0x99, 0x53, 0x92, 0x5a, 0x24,
	0xc1, 0xaa, 0xc0, 0xac, 0xc1, 0x6d, 0x24, 0xa3, 0x07, 0xb2, 0x54, 0x0f, 0x61, 0xa1, 0x9e, 0x1b,
	0x58, 0xda, 0x35, 0xaf, 0xa4, 0xa8, 0x32, 0x08, 0xaa, 0x56, 0xca, 0x92, 0x8b, 0x1b, 0x49, 0x18,
	0x64, 0x5f, 0x76, 0x6a, 0x25, 0xd4, 0x51, 0x20, 0xa6, 0x90, 0x08, 0x17, 0x6b, 0x59, 0x62, 0x4e,
	0x29, 0xcc, 0x59, 0x10, 0x8e, 0x15, 0x93, 0x05, 0x63, 0x12, 0x1b, 0xd8, 0x54, 0x63, 0x40, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xd0, 0x4c, 0xd8, 0x80, 0xe6, 0x00, 0x00, 0x00,
}

func TestMarshalProto(t *testing.T) {
	tp := &TestProto{
		Id:     "uuid",
		Name:   "elvizlai",
		Age:    18,
		NextId: "next-uuid",
		Filter: map[string]string{
			"k1": "v1",
		},
	}
	data := MarshalProto(tp)
	t.Log(string(data))
}

const dataStr = `{"id":"uuid","name":"elvizlai","age":"18","nextId":"next-uuid","filter":{"k1":"v1"}}`

func TestUnmarshalProto(t *testing.T) {
	tp := &TestProto{}

	err := UnmarshalProto([]byte(dataStr), tp)
	t.Log(err, tp)

	buff := bytes.NewBuffer([]byte(dataStr))
	err = UnmarshalProto(buff, tp)
	t.Log(err, tp)
}
