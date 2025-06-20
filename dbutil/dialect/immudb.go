package dialect

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/azhai/allgo/match"
)

const ImmuDBPort uint16 = 3322

// ImmuDB ImmuDB数据库
type ImmuDB struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port,omitempty"`
	Database string `json:"database,omitempty"`
	Sslmode  string `json:"sslmode,omitempty"` // 例如 disable
	Options  `json:"options,omitempty"`
}

// IsSupport 是否支持特性
func (ImmuDB) IsSupport(feature string) bool {
	var lacks = map[string]bool{
		FeatBatchInsert: false,
		FeatForeignKey:  false,
	}
	if without, ok := lacks[feature]; ok {
		return without
	}
	return true
}

// TypeName 驱动名
func (ImmuDB) TypeName() string {
	return "immudb"
}

// ImporterPath 驱动支持库
func (ImmuDB) ImporterPath() string {
	return "github.com/codenotary/immudb/pkg/stdlib"
}

// QuoteIdent 字段或表名脱敏
func (ImmuDB) QuoteIdent(ident string) string {
	return match.WrapWith(ident, `"`, `"`)
}

// GetParamString 获得连接参数
func (d ImmuDB) GetParamString() string {
	d.Options.MakeSize(0)
	if mode := d.Sslmode; mode != "" {
		d.Options.Set("sslmode", mode)
	} else {
		d.Options.Set("sslmode", "disable")
	}
	return d.Options.Encode()
}

// BuildDSN 生成DSN连接串
func (d ImmuDB) BuildDSN() string {
	addr := DefaultHost
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("immudb://%s/%s?", addr, d.Database)
	return dsn + d.GetParamString()
}

// BuildFullDSN 生成带账号的完整DSN
func (d ImmuDB) BuildFullDSN(username, password string) string {
	dsn, head := d.BuildDSN(), "immudb://"
	if strings.HasPrefix(dsn, head) {
		account := GetAccount(username, password)
		dsn = head + account + "@" + dsn[len(head):]
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (ImmuDB) GetCurrentDB(db *sql.DB) string {
	dbname, query := "", `SELECT name FROM DATABASES() LIMIT 1`
	_ = db.QueryRow(query).Scan(&dbname)
	return dbname
}

// FindTableInfos 查找表信息
func (d ImmuDB) FindTableInfos(db *sql.DB) []*TableSchema {
	query := `SELECT name, NULL as comment FROM TABLES() ORDER BY name`
	tables := QueryTableInfos(db, query)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (d ImmuDB) FetchColumnInfos(db *sql.DB, table string) []*ColumnInfo {
	query := `SELECT name, NULL as default, nullable, type, '' as col_type,
max_length, NULL as comment, CASE WHEN "primary" THEN 'primary'
WHEN "unique" THEN 'unique' WHEN indexed THEN 'index' END as col_key,
CASE WHEN "auto_increment" THEN 'auto_incre' END as extra FROM COLUMNS($1)`
	return QueryTableColumns(db, query, table)
}
