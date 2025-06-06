package dialect

import (
	"database/sql"
	"fmt"

	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/match"
)

const MysqlPort uint16 = 3306

func init() {
	dbutil.RegisterDialect(&Mysql{}, "mariadb")
}

// Mysql MySQL或MariaDB数据库
type Mysql struct {
	Host           string `json:"host"`
	Port           uint16 `json:"port,omitempty"`
	Database       string `json:"database,omitempty"`
	dbutil.Options `json:"options,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (Mysql) IsRelationalDB() bool {
	return true
}

// TypeName 驱动名
func (Mysql) TypeName() string {
	return "mysql"
}

// ImporterPath 驱动支持库
func (Mysql) ImporterPath() string {
	return "github.com/go-sql-driver/mysql"
}

// QuoteIdent 字段或表名脱敏
func (Mysql) QuoteIdent(ident string) string {
	return match.WrapWith(ident, "`", "`")
}

// GetParamString 获得连接参数
func (d Mysql) GetParamString() string {
	opts := "parseTime=true&loc=Local"
	if d.Options.MakeSize(0) == 0 {
		return opts
	}
	if !d.Options.Has("parseTime") {
		d.Options.Set("parseTime", "true")
	}
	if !d.Options.Has("loc") {
		d.Options.Set("loc", "Local")
	}
	return d.Options.Encode()
}

// BuildDSN 生成DSN连接串
func (d Mysql) BuildDSN() string {
	addr := dbutil.DefaultHost
	if d.Host != "" {
		addr = dbutil.GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("tcp(%s)/%s?", addr, d.Database)
	return dsn + d.GetParamString()
}

// BuildFullDSN 生成带账号的完整DSN
func (d Mysql) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if dsn != "" {
		account := dbutil.GetAccount(username, password)
		dsn = account + "@" + dsn
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (Mysql) GetCurrentDB(db *sql.DB) string {
	dbname, query := "", `SELECT DATABASE()`
	_ = db.QueryRow(query).Scan(&dbname)
	return dbname
}

// getDbName 获得当前数据库名
func (d Mysql) getDbName(db *sql.DB) string {
	if d.Database == "" {
		d.Database = d.GetCurrentDB(db)
	}
	return d.Database
}

// FindTableInfos 查找表信息
func (d Mysql) FindTableInfos(db *sql.DB) []*dbutil.TableSchema {
	query := `SELECT table_name, table_comment
FROM information_schema.tables WHERE table_schema = ?
AND table_type = 'BASE TABLE' ORDER BY table_name`
	tables := dbutil.QueryTableInfos(db, query, d.getDbName(db))
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (d Mysql) FetchColumnInfos(db *sql.DB, table string) []*dbutil.ColumnInfo {
	query := `SELECT column_name, column_default, is_nullable = 'YES' as is_nullable,
data_type, column_type, character_maximum_length, column_comment, column_key, extra
FROM information_schema.columns WHERE table_name = ? AND table_schema = ?
ORDER BY ordinal_position`
	return dbutil.QueryTableColumns(db, query, table, d.getDbName(db))
}
