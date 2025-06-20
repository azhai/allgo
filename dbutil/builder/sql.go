package builder

// Builder 用于构建SQL语句
type Builder interface {
	// String 构建SQL语句
	String() string
	// Args 返回SQL语句中的参数
	Args() []any
}

type SQL struct {
	string
	args     []any
	isParsed bool
}

func (s *SQL) String() string {
	return s.string
}

func (s *SQL) Args() []any {
	return s.args
}

// ShowBuilder 用于构建SHOW或简单SELECT语句
type ShowBuilder struct {
	head        string // 头部信息，如"SHOW"或"SELECT"
	Information string
	Where       *WhereBuilder
	SQL
}

// InsertBuilder 用于构建INSERT语句
type InsertBuilder struct {
	head      string // 头部信息，如"INSERT INTO"或"REPLACE INTO"
	Returning string // 返回字段，如"RETURNING id"

	Columns []string
	Values  []any
	Others  [][]any

	Conflict
	FromTable
	SQL
}

func Insert() *InsertBuilder {
	return &InsertBuilder{}
}

// DeleteBuilder 用于构建DELETE语句
type DeleteBuilder struct {
	Head   string // 头部信息，如"DELETE FROM"或"UPDATE"或"SELECT"
	Limit  int
	Offset int
	Tail   string // 额外的SQL语句，如"FOR UPDATE"或"LOCK IN SHARE MODE"或"ROLL UP"

	Where *WhereBuilder
	Join  []JoinTable
	FromTable
	SQL
}

// UpdateBuilder 用于构建UPDATE语句
type UpdateBuilder struct {
	Update UpdateSet
	Conflict
	*DeleteBuilder
}

// SelectBuilder 用于构建SELECT语句
// Select键为别名，值为字段与函数列表
type SelectBuilder struct {
	Select  map[string]SelField
	GroupBy string
	Having  string
	OrderBy []string
	*DeleteBuilder
}

// UpdateSet 用于构建UPDATE语句
type UpdateSet struct {
	Keys   []string
	Values []string
	SQL
}

// SelField 用于构建SELECT语句
type SelField struct {
	Col string
	Ops []string
	SQL
}
