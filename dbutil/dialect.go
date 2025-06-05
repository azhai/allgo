package dbutil

import (
	"fmt"
	"strings"
	"sync"

	"github.com/azhai/allgo/match"
)

const DefaultHost = "127.0.0.1"

var (
	dialects = make(map[string]Dialect)
	diaMutex = new(sync.RWMutex)
)

// Dialect 数据库类型
type Dialect interface {
	IsRelationalDB() bool                                    // 是否关系数据库
	TypeName() string                                        // 类型名
	ImporterPath() string                                    // 驱动支持库
	QuoteIdent(ident string) string                          // 字段或表名脱敏
	BuildDSN() string                                        // 生成DSN连接串
	BuildFullDSN(username, password string) string           // 生成带账号的完整DSN
	GetCurrentDB(db *DBServ) string                          // 获得当前数据库名
	FindTableInfos(db *DBServ) []*TableSchema                // 查找表信息
	FetchColumnInfos(db *DBServ, table string) []*ColumnInfo // 查找字段信息
}

// RegisterDialect 注册数据库类型
func RegisterDialect(dialect Dialect, names ...string) {
	diaMutex.Lock()
	defer diaMutex.Unlock()
	dialects[dialect.TypeName()] = dialect
	for _, name := range names {
		dialects[name] = dialect
	}
}

// ParseSchema 解析数据库连接串，获得数据库类型
func ParseSchema(dsn string) string {
	sch := match.Word(dsn).MatchFirstID()
	return strings.ToLower(sch)
}

// GetAddr 获得数据库的连接地址和端口，用于TCP协议的数据库连接
func GetAddr(host string, port uint16) string {
	if port == 0 {
		return host
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// GetAccount 获得账号密码，用于DSN连接串
func GetAccount(username, password string) string {
	if username == "" {
		return password
	}
	return fmt.Sprintf("%s:%s", username, password)
}
