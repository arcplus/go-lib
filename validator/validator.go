package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Validator interface {
	Validate() error
}

// Validate checks constraints
func Validate(req interface{}, constraints ...string) error {
	// using Validate method if exist
	if v, ok := req.(Validator); ok {
		return v.Validate()
	}

	if len(constraints) == 0 {
		return nil
	}

	rv := reflect.Indirect(reflect.ValueOf(req))

	var err error

	for _, c := range constraints {
		if c == "" {
			continue
		}

		nodeName, funcName, funcParams := splitToken(c)

		rv := rv

		for _, node := range strings.Split(nodeName, ".") {
			rv, err = valueWalker(rv, node)
			if err != nil {
				return fmt.Errorf("field '%s' failed with %s", nodeName, err.Error())
			}
		}

		if funcName == "" {
			if !normalValidate(rv) {
				return fmt.Errorf("field '%s' failed with should not empty", nodeName)
			}
		} else {
			if f, ok := funcMap[funcName]; ok && f != nil {
				err = f(rv, funcParams)
				if err != nil {
					return fmt.Errorf("field '%s' failed with %s '%s'", nodeName, err, c)
				}
			} else {
				return fmt.Errorf("field '%s' failed with func '%s' not exist", nodeName, funcName)
			}
		}
	}

	return nil
}

// nodeName funcName funcParams
func splitToken(c string) (string, string, string) {
	i := strings.Index(c, ":")
	if i == -1 {
		return c, "", ""
	}

	j := strings.Index(c[i+1:], "(")
	if j == -1 {
		return c[:i], c[i+1:], ""
	}

	return c[:i], c[i+1 : i+j+1], c[i+j+2 : len(c)-1]
}

func valueWalker(rv reflect.Value, node string) (reflect.Value, error) {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if node == "" {
		return rv, nil
	}

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		i, err := strconv.Atoi(node)
		if err != nil {
			return rv, err
		}

		if i > rv.Len()-1 {
			return rv, errors.New("index out of range")
		}

		return reflect.Indirect(rv.Index(i)), err
	case reflect.Map:
		var keyValue reflect.Value
		for _, v := range rv.MapKeys() {
			if fmt.Sprint(v.Interface()) == node {
				keyValue = v
				break
			}
		}

		if !keyValue.IsValid() {
			return rv, errors.New("map has no key '" + node + "'")
		}

		return reflect.Indirect(rv.MapIndex(keyValue)), nil
	case reflect.Struct:
		rv = rv.FieldByName(underscoreToCamelCase(node))
		if !rv.IsValid() {
			return rv, errors.New("no such field '" + node + "'")
		}
		return reflect.Indirect(rv), nil
	case reflect.Invalid:
		return rv, errors.New("'" + node + "' parent invalid")
	case reflect.Interface:
		return valueWalker(rv.Elem(), node)
	}

	return rv, errors.New("type error '" + node + "'")
}

// abc_xyz to AbcXyz
func underscoreToCamelCase(s string) string {
	// is empty or start with UpperCase, do nothing
	if len(s) == 0 || (s[0] >= 'A' && s[0] <= 'Z') {
		return s
	}

	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}

// ValidateFunc ValidateFunc
type ValidateFunc func(rv reflect.Value, p string) error

var funcMap = map[string]ValidateFunc{
	"range":   _range,
	"len":     _range,
	"in":      in,
	"each":    each,
	"default": defaultValue,
}

// Register add new ValidateFunc
func Register(name string, f ValidateFunc) {
	funcMap[name] = f
}

// [low,high)
func _range(rv reflect.Value, p string) error {
	var t interface{}
	var typeStr string
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		t = rv.Len()
		typeStr = "len"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t = rv.Int()
		typeStr = "num"
	case reflect.Float32, reflect.Float64:
		t = rv.Float()
		typeStr = "float"
	default:
		return fmt.Errorf("func range usage error %s", rv.Kind())
	}

	lh := strings.Split(p, "|")

	switch len(lh) {
	case 2:
		switch typeStr {
		case "len":
			high, err := strconv.Atoi(lh[1])
			if err != nil {
				return err
			}
			if t.(int) >= high {
				return fmt.Errorf("%s '%d' gte", typeStr, t)
			}
		case "num":
			high, err := strconv.ParseInt(lh[1], 10, 64)
			if err != nil {
				return err
			}
			if t.(int64) >= high {
				return fmt.Errorf("%s '%d' gte", typeStr, t)
			}
		case "float":
			high, err := strconv.ParseFloat(lh[1], 64)
			if err != nil {
				return err
			}
			if t.(float64) >= high {
				return fmt.Errorf("%s '%d' gte", typeStr, t)
			}
		}
		fallthrough
	case 1:
		switch typeStr {
		case "len":
			low, err := strconv.Atoi(lh[0])
			if err != nil {
				return err
			}
			if t.(int) >= low {
				return nil
			}
		case "num":
			low, err := strconv.ParseInt(lh[0], 10, 64)
			if err != nil {
				return err
			}
			if t.(int64) >= low {
				return nil
			}
		case "float":
			low, err := strconv.ParseFloat(lh[0], 64)
			if err != nil {
				return err
			}
			if t.(float64) >= low {
				return nil
			}
		}
	}

	return fmt.Errorf("%s '%d' lt", typeStr, t)
}

// in list
func in(rv reflect.Value, p string) error {
	ps := strings.Split(p, "|")

	var v string
	switch rv.Kind() {
	case reflect.String:
		v = rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v = strconv.FormatInt(rv.Int(), 10)
	}

	if v != "" {
		for i := range ps {
			if v == ps[i] {
				return nil
			}
		}
	}

	return fmt.Errorf("'%s' not in", v)
}

// each element
func each(rv reflect.Value, p string) error {
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		l := rv.Len()
		for i := 0; i < l; i++ {
			rv, err := valueWalker(rv.Index(i), p)
			if err != nil {
				return err
			}
			if !normalValidate(rv) {
				return fmt.Errorf("'%s' should not empty", p)
			}
		}
	case reflect.Map:

	}
	return nil
}

// defaultValue set rv to
func defaultValue(rv reflect.Value, p string) error {
	if !rv.CanSet() {
		return errors.New("not settable")
	}

	switch rv.Kind() {
	case reflect.String:
		if rv.String() == "" {
			rv.SetString(p)
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if rv.Int() == 0 {
			v, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				return err
			}
			rv.SetInt(v)
		}
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if rv.Uint() == 0 {
			v, err := strconv.ParseUint(p, 10, 64)
			if err != nil {
				return err
			}
			rv.SetUint(v)
		}
	case reflect.Float64, reflect.Float32:
		if rv.Float() == 0 {
			v, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return err
			}
			rv.SetFloat(v)
		}
	default:
		return errors.New(rv.Kind().String() + " not support now")
	}
	return nil
}

// base type only, no struct should be here.
func normalValidate(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.String:
		return v.String() != ""
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() != 0
	case reflect.Invalid:
		return false
	case reflect.Interface, reflect.Ptr:
		return normalValidate(v.Elem())
	default:
		return true
	}
}
