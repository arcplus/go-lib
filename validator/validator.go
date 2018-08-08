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
	// check if impl Validator interface
	if v, ok := req.(Validator); ok {
		return v.Validate()
	}

	if len(constraints) == 0 {
		return nil
	}

	rv := reflect.Indirect(reflect.ValueOf(req))

	var err error

	for _, c := range constraints {
		nodeName, funcName, funcParams := splitToken(c)

		rv := rv

		nodes := strings.Split(nodeName, ".")
		for i, node := range nodes {
			rv, err = valueWalker(rv, node)
			if err != nil {
				return fmt.Errorf("failed with '%s' %s", strings.Join(nodes[:i+1], "."), err.Error())
			}
		}

		if funcName == "" {
			if !normalValidate(rv) {
				return fmt.Errorf("failed with '%s' should not empty", c)
			}
		} else {
			if f, ok := funcMap[funcName]; ok && f != nil {
				err = f(rv, funcParams)
				if err != nil {
					return fmt.Errorf("failed with %s '%s'", err, c)
				}
			} else {
				return fmt.Errorf("failed with func '%s' not exist", funcName)
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
		return rv, errors.New("parent invalid '" + node + "'")
	case reflect.Interface:
		return valueWalker(rv.Elem(), node)
	}

	return rv, errors.New("type error '" + node + "'")
}

// abc_xyz to AbcXyz
func underscoreToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}

// ValidateFunc ValidateFunc
type ValidateFunc func(rv reflect.Value, p string) error

var funcMap = map[string]ValidateFunc{
	"range": length,
	"in":    in,
	"each":  each,
}

// Register add new ValidateFunc
func Register(name string, f ValidateFunc) {
	funcMap[name] = f
}

// [low,high)
func length(rv reflect.Value, p string) error {
	var t int
	var typeStr string
	switch rv.Kind() {
	case reflect.String, reflect.Map, reflect.Slice, reflect.Array:
		t = rv.Len()
		typeStr = "len"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		t = int(rv.Int())
		typeStr = "num"
	default:
		return fmt.Errorf("func usage error %s", rv.Kind())
	}

	lh := strings.Split(p, "|")

	switch len(lh) {
	case 2:
		high, _ := strconv.Atoi(lh[1])
		if t >= high {
			return fmt.Errorf("%s '%d' gte", typeStr, t)
		}
		fallthrough
	case 1:
		low, err := strconv.Atoi(lh[0])
		if err != nil {
			return err
		}

		if t >= low {
			return nil
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
