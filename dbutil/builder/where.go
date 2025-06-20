package builder

import (
	"bytes"
	"strings"
)

type WhereBuilder struct {
	Or []*Clause
	SQL
}

func (b *WhereBuilder) Parse() *WhereBuilder {
	if b.isParsed {
		return b
	}
	var clauses []string
	for _, v := range b.Or {
		clauses = append(clauses, v.Parse().String())
		b.SQL.args = append(b.SQL.args, v.args...)
	}
	if len(clauses) >= 2 {
		b.SQL.string = "WHERE (" + strings.Join(clauses, ") OR (") + ")"
	} else if len(clauses) == 1 {
		b.SQL.string = "WHERE " + clauses[0]
	}
	b.isParsed = true
	return b
}

// Clause 用于构建WHERE语句
type Clause struct {
	And []string
	SQL
}

func (b *Clause) Where(query string, args ...any) *Clause {
	b.And = append(b.And, query)
	b.SQL.args = append(b.SQL.args, args...)
	return b
}

func (b *Clause) Parse() *Clause {
	if b.isParsed {
		return b
	}
	var buf bytes.Buffer
	for i, v := range b.And {
		if i > 0 {
			buf.WriteString(" AND ")
		}
		buf.WriteString(v)
	}
	b.SQL.string = strings.TrimSpace(buf.String())
	b.isParsed = true
	return b
}
