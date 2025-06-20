package dbutil

import (
	"database/sql"
	"net/url"
	"strings"
	"sync"

	"github.com/azhai/allgo/dbutil/dialect"
	"github.com/azhai/allgo/match"
)

var (
	dialects = make(map[string]Dialect)
	diaMutex = new(sync.RWMutex)
)

func init() {
	RegisterDialect(&dialect.ImmuDB{})
	RegisterDialect(&dialect.Mysql{}, "mariadb")
	RegisterDialect(&dialect.Postgres{}, "pgsql", "postgresql")
	RegisterDialect(&dialect.Redis{}, "dragonfly", "garnet", "keydb", "valkey")
	RegisterDialect(&dialect.Sqlite{}, "sqlite", "limbo", "file")
}

// Dialect 数据库类型
type Dialect interface {
	IsSupport(feature string) bool  // 是否支持特性
	TypeName() string               // 类型名
	ImporterPath() string           // 驱动支持库
	QuoteIdent(ident string) string // 字段或表名脱敏

	BuildDSN() string                              // 生成DSN连接串
	BuildFullDSN(username, password string) string // 生成带账号的完整DSN
	ParseUrlQuery(queryStr string) error           // 解析连接参数
	MergeOptions(optVals url.Values)               // 设置连接参数
	GetParamString() string                        // 获得连接参数

	GetCurrentDB(db *sql.DB) string                                  // 获得当前数据库名
	FindTableInfos(db *sql.DB) []*dialect.TableSchema                // 查找表信息
	FetchColumnInfos(db *sql.DB, table string) []*dialect.ColumnInfo // 查找字段信息
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

// ParseScheme 解析数据库连接串，获得数据库类型
func ParseScheme(dsn string) string {
	sch := match.Word(dsn).MatchFirstID()
	return strings.ToLower(sch)
}
