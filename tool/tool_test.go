package tool

import (
	"bytes"
	"testing"

	"github.com/arcplus/go-lib/internal/pb"
)

func TestMarshalProto(t *testing.T) {
	tp := &pb.TestProto{
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
	tp := &pb.TestProto{}

	err := UnmarshalProto([]byte(dataStr), tp)
	t.Log(err, tp)

	buff := bytes.NewBuffer([]byte(dataStr))
	err = UnmarshalProto(buff, tp)
	t.Log(err, tp)
}
