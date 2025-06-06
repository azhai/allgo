package dbutil

import (
	"database/sql"
	"fmt"
	"maps"
	"net/url"
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
	IsRelationalDB() bool           // 是否关系数据库
	TypeName() string               // 类型名
	ImporterPath() string           // 驱动支持库
	QuoteIdent(ident string) string // 字段或表名脱敏

	BuildDSN() string                              // 生成DSN连接串
	BuildFullDSN(username, password string) string // 生成带账号的完整DSN
	ParseUrlQuery(queryStr string) error           // 解析连接参数
	MergeOptions(optVals url.Values)               // 设置连接参数
	GetParamString() string                        // 获得连接参数

	GetCurrentDB(db *sql.DB) string                          // 获得当前数据库名
	FindTableInfos(db *sql.DB) []*TableSchema                // 查找表信息
	FetchColumnInfos(db *sql.DB, table string) []*ColumnInfo // 查找字段信息
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

// Options 连接参数
type Options struct {
	url.Values
}

// MakeSize 初始化连接参数
func (t *Options) MakeSize(size int) int {
	if t.Values == nil {
		t.Values = make(url.Values, size)
		return size
	}
	return len(t.Values)
}

// ParseUrlQuery 解析连接参数
func (t *Options) ParseUrlQuery(queryStr string) (err error) {
	queryStr = strings.Trim(queryStr, "?&# ")
	if len(queryStr) == 0 {
		t.Values = make(url.Values)
	} else {
		t.Values, err = url.ParseQuery(queryStr)
	}
	return
}

// MergeOptions 合并连接参数
func (t *Options) MergeOptions(optVals url.Values) {
	if t.MakeSize(0) == 0 && len(optVals) > 0 {
		t.Values = maps.Clone(optVals)
		return
	}
	for k, vs := range optVals {
		t.Values[k] = vs
	}
}

// ParseScheme 解析数据库连接串，获得数据库类型
func ParseScheme(dsn string) string {
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

// GetCurrentDBFromDSN 从DSN连接串中获得数据库名
func GetCurrentDBFromDSN(dsn string) string {
	if u, err := url.Parse(dsn); err == nil {
		return strings.Trim(u.Path, "/. ")
	}
	return ""
}
