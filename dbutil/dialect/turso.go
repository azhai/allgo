package dialect

import (
	"database/sql"
	"net/url"
	"regexp"
	"strings"

	"github.com/azhai/allgo/match"
)

// Turso SQLite3数据库
type Turso struct {
	Path    string `json:"path"`
	Options `json:"options,omitempty"`
}

// IsSupport 是否支持特性
func (Turso) IsSupport(feature string) bool {
	var lacks = map[string]bool{
		FeatDollarHolder: false,
		FeatForeignKey:   false,
	}
	if without, ok := lacks[feature]; ok {
		return without
	}
	return true
}

// TypeName 驱动名
func (Turso) TypeName() string {
	return "turso"
}

// ImporterPath 驱动支持库
func (Turso) ImporterPath() string {
	return "github.com/tursodatabase/turso-go"
}

// QuoteIdent 字段或表名脱敏
func (Turso) QuoteIdent(ident string) string {
	return match.WrapWith(ident, "`", "`")
}

// GetParamString 获得连接参数
func (d Turso) GetParamString() string {
	return ""
}

// BuildDSN 生成DSN连接串
func (d Turso) BuildDSN() string {
	var params string
	if params = d.GetParamString(); params != "" {
		params = "?" + params
	}
	return d.Path + params
}

// BuildFullDSN 生成带账号的完整DSN
func (d Turso) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if username != "" {
		dsn += "_auth_user=" + username + "&"
		password = url.QueryEscape(password)
		dsn += "_auth_pass=" + password + "&"
	}
	return dsn
}

// GetCurrentDB 获得当前数据库名
func (Turso) GetCurrentDB(db *sql.DB) string {
	return ""
}

// FindTableInfos 查找表信息
func (d Turso) FindTableInfos(db *sql.DB) []*TableSchema {
	query := `SELECT tbl_name, name FROM sqlite_master WHERE type = 'table'
AND tbl_name NOT LIKE 'sqlite_%' ORDER BY tbl_name`
	tables := QueryTableInfos(db, query)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (Turso) FetchColumnInfos(db *sql.DB, table string) []*ColumnInfo {
	query := `SELECT sql FROM sqlite_master WHERE type = 'table' AND tbl_name = ?`
	var createSQL string
	err := db.QueryRow(query, table).Scan(&createSQL)
	if err != nil || createSQL == "" {
		return nil
	}

	createSQL = strings.ReplaceAll(createSQL, "\n", " ")
	nStart := strings.Index(createSQL, "(")
	nEnd := strings.LastIndex(createSQL, ")")
	reg := regexp.MustCompile(`[^\(,\)]*(\([^\(]*\))?`)
	colCreates := reg.FindAllString(createSQL[nStart+1:nEnd], -1)

	var cols []*ColumnInfo
	pks := make(map[string]bool)
	for _, colStr := range colCreates {
		reg = regexp.MustCompile(`,\s`)
		colStr = reg.ReplaceAllString(colStr, ",")
		colStr = strings.TrimSpace(colStr)
		if strings.HasPrefix(colStr, "PRIMARY KEY") {
			parts := strings.Split(colStr, "(")
			if len(parts) == 2 {
				rightPart := strings.TrimSpace(parts[1])
				pkCols := strings.Split(strings.TrimRight(rightPart, ")"), ",")
				for _, pk := range pkCols {
					pk = strings.Trim(strings.TrimSpace(pk), "`")
					pks[pk] = true
				}
			}
			continue
		}

		col, err := parseString(colStr, pks)
		if err != nil {
			return cols
		}
		cols = append(cols, col)
	}
	return cols
}
