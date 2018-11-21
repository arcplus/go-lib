package mysql

import (
	"bytes"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type sBuilder struct {
	smt   []byte
	table []byte
	expr  []byte
	cond  []byte
	group []byte
	order []byte
	limit []byte
	f     []byte
}

func SELECT(table string) *sBuilder {
	return &sBuilder{
		smt:   []byte("SELECT"),
		table: wrap("FROM " + table),
	}
}

func (b *sBuilder) EXPR(str string) *sBuilder {
	b.expr = wrap(str)
	return b
}

func (b *sBuilder) WHERE(cond string) *sBuilder {
	b.cond = wrap("WHERE " + cond)
	return b
}

func (b *sBuilder) GROUPBYHAVING(str string) *sBuilder {
	b.group = wrap("GROUP BY " + str)
	return b
}

func (b *sBuilder) ORDERBY(str string) *sBuilder {
	b.order = wrap("ORDER BY " + str)
	return b
}

// Limit [offset,] row_count TODO optimize fmt.Sprint
func (b *sBuilder) LIMIT(x ...interface{}) *sBuilder {
	switch len(x) {
	case 1:
		b.limit = wrap("LIMIT " + fmt.Sprint(x[0]))
	case 2:
		b.limit = wrap("LIMIT " + fmt.Sprint(x[0]) + "," + fmt.Sprint(x[1]))
	}
	return b
}

func (b *sBuilder) FOR(str string) *sBuilder {
	b.f = wrap("FOR " + str)
	return b
}

var space = []byte(" ")
var end = []byte(";")

// warp with space
func wrap(str string) []byte {
	if str == "" {
		return nil
	}
	return []byte(" " + str)
}

// Build map[string]interface{}, [bool?hasInStatement]
// gen query, args, error
func (b *sBuilder) Build(args ...interface{}) (string, []interface{}, error) {
	buf := &bytes.Buffer{}

	buf.Write(b.smt)
	buf.Write(b.expr)
	buf.Write(b.table)
	buf.Write(b.cond)
	buf.Write(b.group)
	buf.Write(b.order)
	buf.Write(b.limit)
	buf.Write(b.f)
	buf.Write(end)

	if l := len(args); l != 0 {
		query, args, err := sqlx.Named(buf.String(), args[0])
		if l > 1 {
			query, args, err = sqlx.In(query, args...)
		}
		return query, args, err
	}

	return buf.String(), nil, nil
}
