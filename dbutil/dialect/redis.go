package dialect

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/azhai/allgo/match"
)

const RedisPort uint16 = 6379

// Redis Redis缓存
// Redis URI https://www.iana.org/assignments/uri-schemes/prov/redis
type Redis struct {
	Host     string `json:"host"`
	Port     uint16 `json:"port,omitempty"`
	Database int    `json:"database,omitempty"`
	Options  `json:"options,omitempty"`
}

// IsSupport 是否支持特性
func (Redis) IsSupport(feature string) bool {
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
	return match.WrapWith(ident, "'", "'")
}

// GetParamString 获得连接参数
func (d Redis) GetParamString() string {
	opts := ""
	if d.Options.MakeSize(0) == 0 {
		return opts
	}
	return d.Options.Encode()
}

// BuildDSN 生成DSN连接串
func (d Redis) BuildDSN() string {
	addr := DefaultHost
	if d.Host != "" {
		addr = GetAddr(d.Host, d.Port)
	}
	dsn := fmt.Sprintf("redis://%s/%d?", addr, d.Database)
	return dsn + d.GetParamString()
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
func (Redis) GetCurrentDB(db *sql.DB) string {
	return ""
}

// FindTableInfos 查找表信息
func (Redis) FindTableInfos(db *sql.DB) []*TableSchema {
	return nil
}

// FetchColumnInfos 查找字段信息
func (Redis) FetchColumnInfos(db *sql.DB, table string) []*ColumnInfo {
	return nil
}
