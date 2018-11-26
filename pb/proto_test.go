package pb

import (
	"bytes"
	"testing"

	"github.com/arcplus/go-lib/internal/pb"
)

func TestMarshal(t *testing.T) {
	var tp *pb.TestProto
	data, err := Marshal(tp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))

	tp = &pb.TestProto{
		Id:     "uuid",
		Name:   "elvizlai",
		Age:    18,
		NextId: "next-uuid",
		Filter: map[string]string{
			"k1": "v1",
		},
	}
	data, err = Marshal(tp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

const dataStr = `{"id":"uuid","name":"elvizlai","age":"18","nextId":"next-uuid","filter":{"k1":"v1"}}`

func TestUnmarshal(t *testing.T) {
	tp := &pb.TestProto{}

	err := Unmarshal([]byte(dataStr), tp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tp)

	buff := bytes.NewBuffer([]byte(dataStr))
	err = Unmarshal(buff, tp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tp)

	tp.Reset()

	err = Unmarshal([]byte("null"), tp)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tp == nil)
}

func TestMarshalAny(t *testing.T) {
	list := make([]Message, 3)
	list[1] = &pb.TestProto{}

	as := MarshalAny(list[1])
	t.Log(as, len(as))

	as = MarshalAny(list)
	t.Log(as, len(as))
}
