package tool

import (
	"bytes"
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jbm = &jsonpb.Marshaler{
	EmitDefaults: true,
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
func UnmarshalProto(data []byte, pb proto.Message) error {
	return jbu.Unmarshal(bytes.NewBuffer(data), pb)
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
