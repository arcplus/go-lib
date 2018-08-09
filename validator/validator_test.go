package validator

import (
	"errors"
	"testing"
)

func TestSplitToken(t *testing.T) {
	t.Log(splitToken(""))
	t.Log(splitToken(":uuid"))
	t.Log(splitToken(":uuid()"))
	t.Log(splitToken("xyz:range(1)"))
	t.Log(splitToken("xyz:range(1|2)"))
}

func TestValidateBasic(t *testing.T) {
	var err error
	err = Validate("", "")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)

	err = Validate(0, "")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)

	err = Validate(0.0, "")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)

	var x *int
	err = Validate(x, "")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)
}

func TestValidateSlice(t *testing.T) {
	var err error
	err = Validate([]string{}, "")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)

	err = Validate([]string{""}, "")
	if err != nil {
		t.Fatal("should be not err", err)
	}

	err = Validate([]string{"abc"}, "0")
	if err != nil {
		t.Fatal("should be not err", err)
	}

	err = Validate([]string{"abc"}, "1")
	if err == nil {
		t.Fatal("should be index out of range")
	}
	t.Log(err)

	err = Validate([][]string{{""}}, "0.0")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)
}

func TestValidateMap(t *testing.T) {
	var err error
	err = Validate(map[string]string{}, "")
	if err == nil {
		t.Fatal("should be not empty err")
	}
	t.Log(err)

	err = Validate(map[string]string{}, "xyz")
	if err == nil {
		t.Fatal("should be no such key err")
	}
	t.Log(err)

	err = Validate(map[string]string{"xyz": "abc"}, "xyz")
	if err != nil {
		t.Fatal("should be ok", err)
	}

	err = Validate(map[string]map[string]string{"xyz": {"abc": ""}}, "xyz.m")
	if err == nil {
		t.Fatal("should be no such key err")
	}
	t.Log(err)

	err = Validate(map[string]map[string]string{"xyz": {"abc": ""}}, "xyz.abc")
	if err == nil {
		t.Fatal("should be not empty")
	}
	t.Log(err)

	err = Validate(map[string]map[string]map[string]string{"xyz": {"abc": {"m": "n", "x": ""}}}, "xyz.abc.x")
	if err == nil {
		t.Fatal("should be not empty")
	}
	t.Log(err)
}

func TestValidateStruct(t *testing.T) {
	type A struct {
		Name string
		Age  string
	}

	type B struct {
		A A
	}

	type C struct {
		List []B
	}

	var err error
	a := A{}
	err = Validate(a, "name")
	if err == nil {
		t.Fatal("should be not empty")
	}
	t.Log(err)

	err = Validate(a, "Age")
	if err == nil {
		t.Fatal("should be not empty")
	}
	t.Log(err)

	a.Name = "elvizlai"
	err = Validate(a, "name")
	if err != nil {
		t.Fatal("should be ok", err)
	}

	b := B{A: a}
	err = Validate(b, "a.name")
	if err != nil {
		t.Fatal("should be ok", err)
	}

	err = Validate(b, "a.age")
	if err == nil {
		t.Fatal("should be not empty")
	}
	t.Log(err)

	c := C{List: []B{b}}
	err = Validate(c, "list.0.a.age")
	if err == nil {
		t.Fatal("should be not empty")
	}
	t.Log(err)
}

func TestValidateFunc(t *testing.T) {
	var err error

	err = Validate("x", ":range(2)")
	if err == nil {
		t.Fatal("should be len lt err")
	}
	t.Log(err)

	err = Validate("abcd", ":range(2|4)")
	if err == nil {
		t.Fatal("should be len gte err")
	}
	t.Log(err)

	err = Validate(1, ":range(2|4)")
	if err == nil {
		t.Fatal("should be num lt err")
	}
	t.Log(err)

	err = Validate("xyz", ":in(abc|123|mn)")
	if err == nil {
		t.Fatal("should not in err")
	}
	t.Log(err)

	err = Validate(map[string][]map[string]string{
		"a": {{"b": ""}},
	}, "a:each(b)")
	if err == nil {
		t.Fatal("should be b not empty err")
	}
	t.Log(err)

	type b int32
	type c float32
	type x struct {
		Name string
		Age  int
		B    b
		C    c
		D    *int
	}

	m := x{}
	err = Validate(m, "name:default(elvizlai)", "age:default(18)", "b:default(30)", "c:default(3.14)")
	t.Log(err, m)
}

func TestValidateX(t *testing.T) {
	type A struct {
		Name string
		Age  int
		A    *A
		I    interface{}
	}

	a := A{
		Name: "elvizlai",
		Age:  0,
	}

	var err error

	err = Validate(a, "Age")
	if err == nil {
		t.Fatal("should not empty err")
	}
	t.Log(err)

	err = Validate(a, "A")
	if err == nil {
		t.Fatal("should not empty err")
	}
	t.Log(err)

	err = Validate(a, "A.Name")
	if err == nil {
		t.Fatal("should parent invalid err")
	}
	t.Log(err)

	err = Validate(a, "I")
	if err == nil {
		t.Fatal("should not empty err")
	}
	t.Log(err)

	var x *A
	a.I = x

	err = Validate(a, "I")
	if err == nil {
		t.Fatal("should not empty err")
	}
	t.Log(err)

	a.I = 123

	err = Validate(a, "I")
	if err != nil {
		t.Fatal(err)
	}

	a.I = &a

	err = Validate(a, "I.Name")
	if err != nil {
		t.Fatal(err)
	}

	err = Validate(a, "I.age")
	if err == nil {
		t.Fatal("should not empty err")
	}
	t.Log(err)

	err = Validate(a, "I.x")
	if err == nil {
		t.Fatal("should no such field")
	}
	t.Log(err)

	err = Validate(a, "I.A.name")
	if err == nil {
		t.Fatal("should parent invalid err")
	}
	t.Log(err)

	a.A = &a
	err = Validate(a, "A")
	if err != nil {
		t.Fatal(err)
	}
	err = Validate(a, "A.Name")
	if err != nil {
		t.Fatal(err)
	}

	a.I = a
	err = Validate(a, "I.A.name")
	if err != nil {
		t.Fatal(err)
	}
}

type a struct {
	Name string
}

func (a a) Validate() error {
	if a.Name != "elvizlai" {
		return errors.New("invalid")
	}
	return nil
}

func TestValidateImpl(t *testing.T) {
	a := a{
		Name: "xyz",
	}

	var err error
	err = Validate(a, "")
	if err == nil {
		t.Fatal("should be invalid")
	}
	t.Log(err)

	a.Name = "elvizlai"
	err = Validate(a, "")
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkValidate(b *testing.B) {
	v := map[string][]map[string]string{
		"a": {{"b": "v"}},
	}
	for i := 0; i < b.N; i++ {
		Validate(v, "a.0.b")
	}
}
