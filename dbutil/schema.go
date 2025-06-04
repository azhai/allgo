package dbutil

import (
	"database/sql"
	"fmt"
)

type TableSchema struct {
	TableInfo
	Columns []*ColumnInfo // 字段列表
}

type TableInfo struct {
	Name    string           // 表名
	Comment sql.Null[string] // 表描述
}

type ColumnInfo struct {
	Name     string           // 字段名
	Default  sql.Null[string] // 默认值
	Nullable bool             // 是否允许为空
	DataType string           // 字段类型
	ColType  string           // 字段类型
	Length   sql.Null[int]    // 字段长度
	Comment  sql.Null[string] // 字段描述
	Index    sql.Null[string] // 索引类型
	Extra    sql.Null[string] // 自动递增等选项
}

func QueryTableInfos(db *DBServ, query string, args ...any) (tables []*TableSchema) {
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

func QueryTableColumns(db *DBServ, query string, args ...any) (cols []*ColumnInfo) {
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
