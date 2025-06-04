package dbutil

import (
	"fmt"
	"strings"
)

const RedisPort uint16 = 6379

// Redis Redis缓存
type Redis struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port,omitempty"`
	Database int    `json:"database,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (Redis) IsRelationalDB() bool {
	return false
}

// TypeName 驱动名
func (Redis) TypeName() string {
	return "redis"
}

// ImporterPath 驱动支持库
func (Redis) ImporterPath() string {
	return "github.com/gomodule/redigo/redis"
}

// QuoteIdent 字段或表名脱敏
func (Redis) QuoteIdent(ident string) string {
	return WrapWith(ident, "'", "'")
}

// BuildDSN 生成DSN连接串
func (d Redis) BuildDSN() string {
	addr := DefaultHost
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("redis://%s/%d?", addr, d.Database)
	return dsn
}

// BuildFullDSN 生成带账号的完整DSN
func (d Redis) BuildFullDSN(username, password string) string {
	dsn, head := d.BuildDSN(), "redis://"
	if strings.HasPrefix(dsn, head) {
		account := GetAccount(username, password)
		dsn = head + account + "@" + dsn[len(head):]
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (Redis) GetCurrentDB(db *DBServ) string {
	return ""
}

// FindTableInfos 查找表信息
func (Redis) FindTableInfos(db *DBServ) []*TableSchema {
	return nil
}

// FetchColumnInfos 查找字段信息
func (Redis) FetchColumnInfos(db *DBServ, table string) []*ColumnInfo {
	return nil
}
