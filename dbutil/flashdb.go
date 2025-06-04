package dbutil

import (
	"net/url"
	"strconv"
)

const FlashDBPort uint16 = 8000

// FlashDB 一个golang写的类似redis的缓存
type FlashDB struct {
	Path             string `json:"path,omitempty"`
	EvictionInterval int    `json:"eviction_interval,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (FlashDB) IsRelationalDB() bool {
	return false
}

// TypeName 驱动名
func (FlashDB) TypeName() string {
	return "flashdb"
}

// ImporterPath 驱动支持库
func (FlashDB) ImporterPath() string {
	return "github.com/arriqaaq/flashdb"
}

// QuoteIdent 字段或表名脱敏
func (FlashDB) QuoteIdent(ident string) string {
	return WrapWith(ident, "'", "'")
}

// BuildDSN 生成DSN连接串
func (d FlashDB) BuildDSN() string {
	data := url.Values{}
	data.Set("path", d.Path)
	if d.EvictionInterval > 0 {
		data.Set("eviction_interval", strconv.Itoa(d.EvictionInterval))
	}
	return data.Encode()
}

// GetCurrentDB 获得当前数据库名
func (FlashDB) GetCurrentDB(db *DBServ) string {
	return ""
}

// BuildFullDSN 生成带账号的完整DSN
func (d FlashDB) BuildFullDSN(username, password string) string {
	return d.BuildDSN()
}

// FindTableInfos 查找表信息
func (FlashDB) FindTableInfos(db *DBServ) []*TableSchema {
	return nil
}

// FetchColumnInfos 查找字段信息
func (FlashDB) FetchColumnInfos(db *DBServ, table string) []*ColumnInfo {
	return nil
}
