package dialect

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/match"
)

const ImmuDBPort uint16 = 3322

func init() {
	dbutil.RegisterDialect(&ImmuDB{})
}

// ImmuDB ImmuDB数据库
type ImmuDB struct {
	Host     string     `json:"host"`
	Port     uint16     `json:"port,omitempty"`
	Database string     `json:"database,omitempty"`
	Sslmode  string     `json:"sslmode,omitempty"` // 例如 disable
	Options  url.Values `json:"options,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (ImmuDB) IsRelationalDB() bool {
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
func (d ImmuDB) BuildDSN() string {
	addr := dbutil.DefaultHost
	if d.Host != "" {
		addr = dbutil.GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("immudb://%s/%s?", addr, d.Database)
	return dsn + d.GetParamString()
}

// BuildFullDSN 生成带账号的完整DSN
func (d ImmuDB) BuildFullDSN(username, password string) string {
	dsn, head := d.BuildDSN(), "immudb://"
	if strings.HasPrefix(dsn, head) {
		account := dbutil.GetAccount(username, password)
		dsn = head + account + "@" + dsn[len(head):]
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (ImmuDB) GetCurrentDB(db *dbutil.DBServ) string {
	_ = db.QueryRow("SELECT CURRENT_DATABASE()").Scan(&db.Name)
	return db.Name
}

// FindTableInfos 查找表信息
func (d ImmuDB) FindTableInfos(db *dbutil.DBServ) []*dbutil.TableSchema {
	query := `SELECT table_name, table_comment
FROM information_schema.tables WHERE table_schema = ?
AND table_type = 'BASE TABLE' ORDER BY table_name`
	if db.Name == "" {
		d.GetCurrentDB(db)
	}
	tables := dbutil.QueryTableInfos(db, query, db.Name)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (d ImmuDB) FetchColumnInfos(db *dbutil.DBServ, table string) []*dbutil.ColumnInfo {
	query := `SELECT column_name, column_default, is_nullable = 'YES' as is_nullable,
data_type, column_type, character_maximum_length, column_comment, column_key, extra
FROM information_schema.columns WHERE table_name = ? AND table_schema = ?
ORDER BY ordinal_position`
	if db.Name == "" {
		d.GetCurrentDB(db)
	}
	return dbutil.QueryTableColumns(db, query, table, db.Name)
}
