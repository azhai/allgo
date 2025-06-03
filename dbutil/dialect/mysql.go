package dialect

import (
	"fmt"

	"github.com/azhai/allgo/dbutil"
)

const MysqlPort uint16 = 3306

// Mysql MySQL或MariaDB数据库
type Mysql struct {
	Host     string `hcl:"host" json:"host"`
	Port     uint16 `hcl:"port,optional" json:"port,omitempty"`
	Database string `hcl:"database,optional" json:"database,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (Mysql) IsRelationalDB() bool {
	return true
}

// Name 驱动名
func (Mysql) Name() string {
	return "mysql"
}

// ImporterPath 驱动支持库
func (Mysql) ImporterPath() string {
	return "github.com/go-sql-driver/mysql"
}

// QuoteIdent 字段或表名脱敏
func (Mysql) QuoteIdent(ident string) string {
	return WrapWith(ident, "`", "`")
}

// BuildDSN 生成DSN连接串
func (d Mysql) BuildDSN() string {
	addr := DefaultHost
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("tcp(%s)/%s", addr, d.Database)
	dsn += "?parseTime=true&loc=Local&"
	return dsn
}

// BuildFullDSN 生成带账号的完整DSN
func (d Mysql) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if dsn != "" {
		account := GetAccount(username, password)
		dsn = account + "@" + dsn
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (Mysql) GetCurrentDB(db *dbutil.DBServ) string {
	_ = db.QueryRow("SELECT DATABASE()").Scan(&db.Name)
	return db.Name
}

// FindTableInfos 查找表信息
func (d Mysql) FindTableInfos(db *dbutil.DBServ) []*TableSchema {
	query := `SELECT table_name, table_comment
FROM information_schema.tables WHERE table_schema = ?
AND TABLE_TYPE = 'BASE TABLE' AND ENGINE IN ('MyISAM', 'InnoDB', 'TokuDB')`
	tables := QueryTableInfos(db, query)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (d Mysql) FetchColumnInfos(db *dbutil.DBServ, table string) []*ColumnInfo {
	query := `SELECT column_name, column_default, is_nullable = 'YES' as is_nullable,
data_type, column_type, character_maximum_length, column_comment, column_key, extra
FROM information_schema.columns WHERE table_name = ? AND table_schema = ?
ORDER BY ordinal_position`
	if db.Name == "" {
		d.GetCurrentDB(db)
	}
	return QueryTableColumns(db, query, table, db.Name)
}
