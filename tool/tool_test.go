package tool

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"

	"github.com/arcplus/go-lib/internal/pb"
)

func TestMarshalProto(t *testing.T) {
	var tp *pb.TestProto
	data := MarshalProto(tp)
	t.Log((data))

	tp = &pb.TestProto{
		Id:     "uuid",
		Name:   "elvizlai",
		Age:    18,
		NextId: "next-uuid",
		Filter: map[string]string{
			"k1": "v1",
		},
	}
	data = MarshalProto(tp)
	t.Log(string(data))
}

const dataStr = `{"id":"uuid","name":"elvizlai","age":"18","nextId":"next-uuid","filter":{"k1":"v1"}}`

func TestUnmarshalProto(t *testing.T) {
	tp := &pb.TestProto{}

	err := UnmarshalProto([]byte(dataStr), tp)
	t.Log(err, tp)

	buff := bytes.NewBuffer([]byte(dataStr))
	err = UnmarshalProto(buff, tp)
	t.Log(err, tp)

	tp.Reset()

	err = UnmarshalProto([]byte{}, tp)
	t.Log(err, tp)
}

func TestProtoToAnyX(t *testing.T) {
	list := make([]proto.Message, 3)

	list[1] = &pb.TestProto{}

	as := ProtoToAnyX(list)

	t.Log(as)
}
