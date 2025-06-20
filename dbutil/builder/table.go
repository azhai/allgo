package builder

import (
	"bytes"
	"regexp"
	"strings"
)

// FromTable 用于构建FROM语句
type FromTable struct {
	Table string
	Alias string
	SQL
}

func From(query string) FromTable {
	return FromTable{SQL: SQL{query, nil, false}}
}

func (b FromTable) Parse() FromTable {
	if b.isParsed {
		return b
	}
	re := regexp.MustCompile(`(?i)\s*(\w+)\s+(AS\s+(\w+))?`)
	matches := re.FindStringSubmatch(b.SQL.string)
	if len(matches) <= 1 {
		return b
	}
	b.Table = matches[1]
	if len(matches) > 3 {
		b.Alias = matches[3]
		b.SQL.string = b.Table + " AS " + b.Alias
	}
	b.isParsed = true
	return b
}

// JoinTable 用于构建JOIN语句
type JoinTable struct {
	Type string
	On   *WhereBuilder
	FromTable
	SQL
}

func Join(query string, args ...any) *JoinTable {
	return &JoinTable{SQL: SQL{query, args, false}}
}

func (b *JoinTable) Parse() *JoinTable {
	if b.isParsed {
		return b
	}
	re := regexp.MustCompile(`(?i)((\w+\s+)?JOIN\s+)?(\w+)\s+(AS\s+(\w+))?ON\s+(.+)`)
	matches := re.FindStringSubmatch(b.SQL.string)
	if len(matches) <= 1 {
		return b
	}
	b.Type = strings.ToUpper(matches[2])
	table := matches[3] + " " + matches[4]
	b.FromTable = From(table).Parse()
	var buf bytes.Buffer
	buf.WriteString(b.Type)
	buf.WriteString(" JOIN ")
	buf.WriteString(b.FromTable.String())
	buf.WriteString(" ON ")
	buf.WriteString(b.On.String())
	b.SQL.string = strings.TrimSpace(buf.String())
	b.SQL.args = append(b.SQL.args, b.On.Args()...)
	b.isParsed = true
	return b
}

func (b *JoinTable) String() string {
	var join = "JOIN"
	if b.Type != "" {
		join = b.Type + " JOIN"
	}
	return join + " " + b.FromTable.String()
}

// Conflict 用于构建ON CONFLICT语句
type Conflict struct {
	Target []string
	Action string
	SQL
}
