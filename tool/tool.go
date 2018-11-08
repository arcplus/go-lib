package tool

import (
	"bytes"
	"io"
	"reflect"

	"github.com/arcplus/go-lib/errs"
	"github.com/arcplus/go-lib/json"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

var jbm = &jsonpb.Marshaler{
	EmitDefaults: true,
	OrigName:     true,
}

// MarshalProto convert proto to bytes without error
func MarshalProto(pb proto.Message) []byte {
	buff := &bytes.Buffer{}
	jbm.Marshal(buff, pb)
	return buff.Bytes()
}

var jbu = &jsonpb.Unmarshaler{
	AllowUnknownFields: true,
}

// UnmarshalProto convert bytes to proto
// in should be io.Reader、 bytes or string
func UnmarshalProto(in interface{}, pb proto.Message) error {
	var reader io.Reader
	switch t := in.(type) {
	case []byte:
		if len(t) == 0 {
			t = []byte("null")
		}
		reader = bytes.NewBuffer(t)
	case string:
		if t == "" {
			t = "null"
		}
		reader = bytes.NewBufferString(t)
	case io.Reader:
		reader = t
	default:
		return errs.New(errs.CodeInternal, "in should be io.Reader or bytes")
	}
	return jbu.Unmarshal(reader, pb)
}

var jbmIndent = &jsonpb.Marshaler{
	EmitDefaults: true,
	Indent:       "  ",
}

// MarshalToString convert proto or struct to json string
func MarshalToString(v interface{}, withIndent ...bool) string {
	if v == nil {
		return "nil"
	}

	switch t := v.(type) {
	case proto.Message:
		var marshaler *jsonpb.Marshaler
		if len(withIndent) != 0 && withIndent[0] {
			marshaler = jbmIndent
		} else {
			marshaler = jbm
		}
		s, err := marshaler.MarshalToString(t)
		if err != nil {
			return err.Error()
		}
		return s
	default:
		var data []byte
		var err error
		if len(withIndent) != 0 && withIndent[0] {
			data, err = json.MarshalIndent(v, "", "  ")
		} else {
			data, err = json.Marshal(v)
		}
		if err != nil {
			return err.Error()
		}
		return string(data)
	}
}

// ProtoToAnyX convert []proto.Message to []*any.Any
func ProtoToAnyX(p interface{}) []*any.Any {
	rv := reflect.ValueOf(p)

	switch rv.Kind() {
	case reflect.Slice:
		l := rv.Len()
		as := make([]*any.Any, l)
		for i := 0; i < l; i++ {
			pm, ok := rv.Index(i).Interface().(proto.Message)
			if ok {
				as[i], _ = ptypes.MarshalAny(pm)
			}
		}
		return as
	default:
		return nil
	}
}
