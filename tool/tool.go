package tool

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/arcplus/go-lib/pool"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var jbm = &jsonpb.Marshaler{
	EmitDefaults: true,
}

var jbmIndent = &jsonpb.Marshaler{
	EmitDefaults: true,
	Indent:       "  ",
}

// MarshalToString convert proto or struct to json string
func MarshalToString(v interface{}, withIndent ...bool) string {
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

// MarshalProto convert proto to bytes
func MarshalProto(pb proto.Message) []byte {
	buff := &bytes.Buffer{}
	jbm.Marshal(buff, pb)
	return buff.Bytes()
}

var bufferPool = pool.NewBytesPool()

var stacktracePool = sync.Pool{
	New: func() interface{} {
		return newProgramCounters(64)
	},
}

type programCounters struct {
	pcs []uintptr
}

func newProgramCounters(size int) *programCounters {
	return &programCounters{make([]uintptr, size)}
}
