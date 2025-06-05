package dialect

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/match"
)

const PgsqlPort uint16 = 5432

func init() {
	dbutil.RegisterDialect(&Postgres{}, "pgsql", "postgresql")
}

// Postgres PostgreSQL数据库
type Postgres struct {
	Host     string     `json:"host"`
	Port     uint16     `json:"port,omitempty"`
	Database string     `json:"database,omitempty"`
	Sslmode  string     `json:"sslmode,omitempty"` // 例如 disable
	Options  url.Values `json:"options,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (Postgres) IsRelationalDB() bool {
	return true
}

// TypeName 驱动名
func (Postgres) TypeName() string {
	return "postgres"
}

// ImporterPath 驱动支持库
func (Postgres) ImporterPath() string {
	return "github.com/lib/pq"
}

// QuoteIdent 字段或表名脱敏
func (Postgres) QuoteIdent(ident string) string {
	return match.WrapWith(ident, `"`, `"`)
}

// GetParamString 获得连接参数
func (d Postgres) GetParamString() string {
	if d.Options == nil {
		d.Options = make(url.Values)
	}
	if mode := d.Sslmode; mode != "" {
		d.Options.Set("sslmode", mode)
	} else {
		d.Options.Set("sslmode", "disable")
	}
	return d.Options.Encode()
}

// BuildDSN 生成DSN连接串
func (d Postgres) BuildDSN() string {
	addr := dbutil.DefaultHost
	if d.Host != "" {
		addr = dbutil.GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("postgres://%s/%s?", addr, d.Database)
	return dsn + d.GetParamString()
}

// BuildFullDSN 生成带账号的完整DSN
func (d Postgres) BuildFullDSN(username, password string) string {
	dsn, head := d.BuildDSN(), "postgres://"
	if strings.HasPrefix(dsn, head) {
		account := dbutil.GetAccount(username, password)
		dsn = head + account + "@" + dsn[len(head):]
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (Postgres) GetCurrentDB(db *dbutil.DBServ) string {
	_ = db.QueryRow("SELECT CURRENT_DATABASE()").Scan(&db.Name)
	return db.Name
}

// FindTableInfos 查找表信息
func (d Postgres) FindTableInfos(db *dbutil.DBServ) []*dbutil.TableSchema {
	query := `SELECT c.relname as table_name, d.description as table_comment
FROM pg_catalog.pg_class c 
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
LEFT JOIN pg_catalog.pg_description d ON d.objoid = c.oid AND d.objsubid = 0
WHERE n.nspname = 'public' AND c.relkind = 'r' ORDER BY c.relname`
	tables := dbutil.QueryTableInfos(db, query)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (Postgres) FetchColumnInfos(db *dbutil.DBServ, table string) []*dbutil.ColumnInfo {
	query := `SELECT s.column_name, s.column_default, s.is_nullable = 'YES' as is_nullable,
s.data_type, s.udt_name as column_type, s.character_maximum_length,
d.description as column_comment, p.contype as column_key, p.conname as extra
FROM information_schema.columns s
JOIN pg_catalog.pg_class c ON c.relname=s.table_name AND c.relkind = 'r'
LEFT JOIN pg_catalog.pg_description d ON d.objoid = c.oid AND d.objsubid = s.ordinal_position
LEFT JOIN pg_catalog.pg_constraint p ON p.conrelid = c.oid AND s.ordinal_position = ANY (p.conkey)
WHERE s.table_schema = 'public' AND s.table_name = $1 ORDER BY s.ordinal_position`
	return dbutil.QueryTableColumns(db, query, table)
}
