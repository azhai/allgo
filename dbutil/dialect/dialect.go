package dialect

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/match"
)

const DefaultHost = "127.0.0.1"

var (
	WrapWith = match.WrapWith
	Escape   = url.QueryEscape
)

// Dialect 不同数据库的驱动配置
type Dialect interface {
	IsRelationalDB() bool                                           // 是否关系数据库
	Name() string                                                   // 驱动名
	ImporterPath() string                                           // 驱动支持库
	QuoteIdent(ident string) string                                 // 字段或表名脱敏
	BuildDSN() string                                               // 生成DSN连接串
	BuildFullDSN(username, password string) string                  // 生成带账号的完整DSN
	GetCurrentDB(db *dbutil.DBServ) string                          // 获得当前数据库名
	FindTableInfos(db *dbutil.DBServ) []*TableSchema                // 查找表信息
	FetchColumnInfos(db *dbutil.DBServ, table string) []*ColumnInfo // 查找字段信息
}

// ConnConfig 连接配置
type ConnConfig struct {
	DSN      string     `hcl:"dsn,optional" json:"dsn,omitempty"`
	Type     string     `hcl:"type,label" json:"type"` // 数据库类型
	Key      string     `hcl:"key,label" json:"key"`   // 数据库连接名
	Username string     `hcl:"username,optional" json:"username,omitempty"`
	Password string     `hcl:"password,optional" json:"password,omitempty"`
	Options  url.Values `hcl:"options,optional" json:"options,omitempty"`
	Dialect  Dialect
}

// LoadDialect 加载数据库驱动配置
func (c *ConnConfig) LoadDialect() Dialect {
	if c.Type == "" || c.Dialect != nil {
		return c.Dialect
	}
	c.Dialect = CreateDialectByName(c.Type)
	return c.Dialect
}

// Name 数据库驱动名
func (c *ConnConfig) Name() string {
	if d := c.LoadDialect(); d != nil {
		return d.Name()
	}
	return c.Type
}

// GetDSN 获取DSN连接串，可选是否带账号密码
func (c *ConnConfig) GetDSN(full bool) string {
	var dsn string
	if d := c.LoadDialect(); d != nil {
		if c.DSN == "" {
			c.DSN = d.BuildDSN()
		}
		if full {
			dsn = d.BuildFullDSN(c.Username, c.Password)
		}
	}
	if dsn == "" {
		dsn = c.DSN
	}
	if args := c.Options.Encode(); args != "" {
		dsn += args
	}
	return strings.TrimRight(dsn, " ?&")
}

// CreateDialectByName 根据名称创建驱动配置
func CreateDialectByName(name string) Dialect {
	name = strings.ToLower(name)
	switch name {
	default:
		return nil
	case "flashdb":
		return &FlashDB{}
	case "mariadb", "mysql":
		return &Mysql{}
	case "pgsql", "postgres":
		return &Postgres{}
	case "redis":
		return &Redis{}
	case "sqlite", "sqlite3":
		return &Sqlite{}
	}
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
