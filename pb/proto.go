package pb

import (
	"bytes"
	"errors"
	"io"
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

var (
	pbm = &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
	}

	pbu = &jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}
)

// alias
type (
	Message = proto.Message
)

// Marshal marshals a protocol buffer into JSON.
// caution: nil proto.Message returns null with no error
func Marshal(pb Message) ([]byte, error) {
	buff := &bytes.Buffer{}
	err := pbm.Marshal(buff, pb)
	if err != nil && err.Error() == "Marshal called with nil" {
		return []byte("null"), nil
	}
	return buff.Bytes(), err
}

// MustMarshal convert proto to JSON without error.
func MustMarshal(pb Message) []byte {
	data, _ := Marshal(pb)
	return data
}

// Unmarshal unmarshals a JSON object stream into a protocol buffer.
// caution: should handle null manually
func Unmarshal(in interface{}, pb Message) error {
	var reader io.Reader
	switch t := in.(type) {
	case []byte:
		reader = bytes.NewBuffer(t)
	case string:
		reader = bytes.NewBufferString(t)
	case io.Reader:
		reader = t
	default:
		return errors.New("type not support")
	}

	return pbu.Unmarshal(reader, pb)
}

// MarshalAny takes the protocol buffer and encodes it into google.protobuf.Any.
// it support []proto.Message to []*any.Any or proto.Message to [0]*any.Any
func MarshalAny(p interface{}) []*any.Any {
	// check if is proto
	if v, ok := p.(Message); ok {
		a, _ := ptypes.MarshalAny(v)
		return []*any.Any{a}
	}

	rv := reflect.ValueOf(p)

	switch rv.Kind() {
	case reflect.Slice:
		l := rv.Len()
		as := make([]*any.Any, l)
		for i := 0; i < l; i++ {
			pm, ok := rv.Index(i).Interface().(Message)
			if ok {
				as[i], _ = ptypes.MarshalAny(pm)
			}
		}
		return as
	default:
		return nil
	}
}
