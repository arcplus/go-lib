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
	expr  []byte
	table []byte
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

// INSERT [LOW_PRIORITY | DELAYED | HIGH_PRIORITY] [IGNORE]
//     [INTO] tbl_name
//     insert_values
//     [ON DUPLICATE KEY UPDATE assignment_list]

// insert_values:
//     [(col_name [, col_name] ...)]
//     {VALUES | VALUE} (expr_list) [, (expr_list)] ...
// |   SET assignment_list
// |   [(col_name [, col_name] ...)]
//     SELECT ...

// expr_list:
//     expr [, expr] ...

// assignment:
//     col_name = expr

// assignment_list:
//     assignment [, assignment] ...

type iBuilder struct {
	smt   []byte
	pri   []byte
	ig    []byte
	table []byte
	cls   []string
	vls   []string
	dup   []byte
	args  []interface{}
}

func INSERT(table string) *iBuilder {
	return &iBuilder{
		smt:   []byte("INSERT"),
		table: wrap("INTO " + table),
	}
}

func (b *iBuilder) HIGHPRIORITY() *iBuilder {
	b.pri = wrap("HIGH_PRIORITY")
	return b
}

func (b *iBuilder) LOWPRIORITY() *iBuilder {
	b.pri = wrap("LOW_PRIORITY")
	return b
}

func (b *iBuilder) IGNORE() *iBuilder {
	b.ig = wrap("IGNORE")
	return b
}

func (b *iBuilder) COLUMN(cls []string) *iBuilder {
	b.cls = cls
	return b
}

func (b *iBuilder) VALUES(vls []string) *iBuilder {
	b.vls = vls
	return b
}

func (b *iBuilder) COLUMNVALUES(cv map[string]interface{}) *iBuilder {
	l := len(cv)
	b.cls = make([]string, l)
	b.args = make([]interface{}, l)

	i := 0
	for k, v := range cv {
		b.cls[i] = k
		b.args[i] = v
		i++
	}

	return b
}

func (b *iBuilder) ONDUPLICATE(cmd string) *iBuilder {
	b.dup = wrap("ON DUPLICATE " + cmd)
	return b
}

func (b *iBuilder) Build(args ...interface{}) (string, []interface{}, error) {
	buf := &bytes.Buffer{}

	buf.Write(b.smt)
	buf.Write(b.pri)
	buf.Write(b.ig)
	buf.Write(b.table)
	buf.Write(wrap("(" + strings.Join(b.cls, ", ") + ")"))

	if b.vls == nil {
		l := len(b.cls)
		b.vls = make([]string, l)
		for i := 0; i < l; i++ {
			b.vls[i] = "?"
		}
	}

	buf.Write(wrap("VALUES (" + strings.Join(b.vls, ", ") + ")"))
	buf.Write(b.dup)
	buf.Write(end)

	if b.args != nil {
		return buf.String(), b.args, nil
	}

	if l := len(args); l != 0 {
		query, args, err := sqlx.Named(buf.String(), args[0])
		if l > 1 {
			query, args, err = sqlx.In(query, args...)
		}
		return query, args, err
	}

	return buf.String(), nil, nil
}

// TODO
// UPDATE [LOW_PRIORITY] [IGNORE] table_reference
//     SET assignment_list
//     [WHERE where_condition]
//     [ORDER BY ...]
//     [LIMIT row_count]

// assignment:
//     col_name = value

// assignment_list:
//     assignment [, assignment] ...
