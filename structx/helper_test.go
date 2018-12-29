package structx

import (
	"reflect"
	"testing"
)

func TestMapKeys(t *testing.T) {
	var m map[string]interface{}

	result1 := MapKeys(m)

	if result1 != nil {
		t.Fatal("result1 should be nil", result1)
	}

	m = map[string]interface{}{
		"a": 1,
	}

	result2 := MapKeys(m)

	if !reflect.DeepEqual(result2, []string{"a"}) {
		t.Fatal("result2 must be []string{a}")
	}

	m["b"] = 2
	m["c"] = 2

	if len(MapKeys(&m).([]string)) != 3 {
		t.Fatal("m len should be 3")
	}

	n := map[int]interface{}{
		1: "a",
	}

	result3 := MapKeys(n)

	if !reflect.DeepEqual(result3, []int{1}) {
		t.Fatal("result3 must be []int{1}")
	}
}
