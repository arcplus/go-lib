package json

import (
	"io"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalToString
func MarshalToString(v interface{}) (string, error) {
	return json.MarshalToString(v)
}

// MarshalIndent
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// UnmarshalFromString
func UnmarshalFromString(str string, v interface{}) error {
	return json.UnmarshalFromString(str, v)
}

// NewEncoder
func NewEncoder(w io.Writer) *jsoniter.Encoder {
	return json.NewEncoder(w)
}

// NewDecoder
func NewDecoder(r io.Reader) *jsoniter.Decoder {
	return json.NewDecoder(r)
}

// Valid
func Valid(data []byte) bool {
	return json.Valid(data)
}

// Get takes interface{} as path.
// If string, it will lookup json map.
// If int, it will lookup json array.
// If '*', it will map to each element of array or each key of map.
func Get(data []byte, path ...interface{}) jsoniter.Any {
	return json.Get(data, path...)
}
