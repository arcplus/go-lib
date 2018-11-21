package mysql

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// SELECT
//     [ALL | DISTINCT | DISTINCTROW ]
//       [HIGH_PRIORITY]
//       [STRAIGHT_JOIN]
//       [SQL_CACHE | SQL_NO_CACHE] [SQL_CALC_FOUND_ROWS]
//     select_expr [, select_expr ...]
//     [FROM table_references
//     [WHERE where_condition]
//     [GROUP BY {col_name | expr | position}
//       [ASC | DESC], ...]
//     [HAVING where_condition]
//     [ORDER BY {col_name | expr | position}
//       [ASC | DESC], ...]
//     [LIMIT {[offset,] row_count | row_count OFFSET offset}]
//     [FOR UPDATE | LOCK IN SHARE MODE]]

type sBuilder struct {
	smt   []byte
	table []byte
	expr  []byte
	cond  []byte
	group []byte
	order []byte
	limit []byte
	f0r   []byte
}

func SELECT(table string) *sBuilder {
	return &sBuilder{
		smt:   []byte("SELECT"),
		table: wrap("FROM " + table),
	}
}

// EXPR x should be string or []string
func (b *sBuilder) EXPR(x interface{}) *sBuilder {
	switch t := x.(type) {
	case string:
		b.expr = wrap(t)
	case []string:
		b.expr = wrap(strings.Join(t, ", "))
	}

	return b
}

// TODO optimize
func reIndent(str string) string {
	buf := &bytes.Buffer{}

	for br := bufio.NewReader(bytes.NewBufferString(str)); ; {
		line, _, err := br.ReadLine()
		if err != nil {
			break
		}

		line = bytes.TrimSpace(line)
		if len(line) != 0 {
			buf.WriteByte(' ')
			buf.Write(line)
		}
	}

	return buf.String()
}

func (b *sBuilder) WHERE(cond string) *sBuilder {
	b.cond = wrap("WHERE" + reIndent(cond))
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
	b.f0r = wrap("FOR " + str)
	return b
}

// warp with space
func wrap(str string) []byte {
	if str == "" {
		return nil
	}
	return []byte(" " + str)
}

var end = []byte(";")

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
	buf.Write(b.f0r)
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
