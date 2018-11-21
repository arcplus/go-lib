package mysql

import (
	"testing"
)

func Test_reIndent(t *testing.T) {
	query := `
	
	id=1
	AND name='abc' 
	AND xyz=123`
	t.Log(reIndent(query))
}

func Benchmark_reIndent(b *testing.B) {
	query := `
	id=1
	AND name='abc' 
	AND xyz=123`
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reIndent(query)
	}
}

func TestFrom(t *testing.T) {
	arg := map[string]interface{}{
		"published": true,
		"authors":   []int{8, 19, 32, 44},
	}

	query, args, err := SELECT("users").EXPR("*").WHERE("id!='' AND published=:published").GROUPBYHAVING("id HAVING id IN (:authors)").ORDERBY("id DESC").LIMIT(10, 10).FOR("UPDATE").Build(arg, true)
	t.Log(query, args, err)
}
