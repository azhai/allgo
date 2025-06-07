package dialect

import (
	"fmt"
	"maps"
	"net/url"
	"strings"
)

const DefaultHost = "127.0.0.1"

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
