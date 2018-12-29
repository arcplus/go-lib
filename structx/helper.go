package structx

import (
	"reflect"
)

// MapKeys return map keys with corresponding type
// map[string]xxx returns []string if map is not zero.
func MapKeys(i interface{}) interface{} {
	v := reflect.Indirect(reflect.ValueOf(i))
	if v.Kind() != reflect.Map {
		return nil
	}

	keys := v.MapKeys()

	keysLen := len(keys)
	if keysLen == 0 {
		return nil
	}

	// gen slice
	v = reflect.MakeSlice(reflect.SliceOf(keys[0].Type()), 0, keysLen)

	// append value
	v = reflect.Append(v, keys...)

	return v.Interface()
}
