package tool

import (
	"encoding/json"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jbm = &jsonpb.Marshaler{
	EmitDefaults: true,
}

// MarshalToString convert proto or struct to json string
func MarshalToString(v interface{}) string {
	switch t := v.(type) {
	case proto.Message:
		s, err := jbm.MarshalToString(t)
		if err != nil {
			return err.Error()
		}
		return s
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return err.Error()
		}
		return string(data)
	}
}
