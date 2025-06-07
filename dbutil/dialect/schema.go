package dialect

import (
	"database/sql"
	"fmt"
)

// TableSchema 表结构
type TableSchema struct {
	Columns   []*ColumnInfo `json:"columns"` // 字段列表
	TableInfo `json:",inline"`
}

// TableInfo 表名与注释
type TableInfo struct {
	Name    string           `json:"name"`    // 表名
	Comment sql.Null[string] `json:"comment"` // 表描述
}

// ColumnInfo 字段信息
type ColumnInfo struct {
	Name     string           `json:"name"`      // 字段名
	Default  sql.Null[string] `json:"default"`   // 默认值
	Nullable bool             `json:"nullable"`  // 是否允许为空
	DataType string           `json:"data_type"` // 字段类型
	ColType  string           `json:"col_type"`  // 字段类型
	Length   sql.Null[int]    `json:"length"`    // 字段长度
	Comment  sql.Null[string] `json:"comment"`   // 字段描述
	Index    sql.Null[string] `json:"index"`     // 索引类型
	Extra    sql.Null[string] `json:"extra"`     // 自动递增等选项
}

// QueryTableInfos 查询表结构
func QueryTableInfos(db *sql.DB, query string, args ...any) (tables []*TableSchema) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var table TableSchema
		err = rows.Scan(&table.Name, &table.Comment)
		if err != nil {
			return
		}
		tables = append(tables, &table)
	}
	return
}

// QueryTableColumns 查询表字段信息
func QueryTableColumns(db *sql.DB, query string, args ...any) (cols []*ColumnInfo) {
	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println("QUERY ERROR:", query, args)
		panic(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var column ColumnInfo
		err = rows.Scan(&column.Name, &column.Default, &column.Nullable, &column.DataType,
			&column.ColType, &column.Length, &column.Comment, &column.Index, &column.Extra)
		if err != nil {
			panic(err)
			return
		}
		cols = append(cols, &column)
	}
	return
}
