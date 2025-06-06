package dialect

import (
	"database/sql"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/match"
)

func init() {
	dbutil.RegisterDialect(&Sqlite{}, "sqlite", "limbo")
}

// Sqlite SQLite3数据库
type Sqlite struct {
	Path           string `json:"path"`
	dbutil.Options `json:"options,omitempty"`
}

// IsRelationalDB 是否关系数据库
func (Sqlite) IsRelationalDB() bool {
	return true
}

// TypeName 驱动名
func (Sqlite) TypeName() string {
	return "sqlite3"
}

// ImporterPath 驱动支持库
func (Sqlite) ImporterPath() string {
	return "github.com/mattn/go-sqlite3"
}

// QuoteIdent 字段或表名脱敏
func (Sqlite) QuoteIdent(ident string) string {
	return match.WrapWith(ident, "`", "`")
}

// GetParamString 获得连接参数
func (d Sqlite) GetParamString() string {
	opts := "cache=shared"
	if d.Options.MakeSize(0) == 0 {
		return opts
	}
	if !d.Options.Has("cache") {
		d.Options.Set("cache", "shared")
	}
	return d.Options.Encode()
}

// BuildDSN 生成DSN连接串
func (d Sqlite) BuildDSN() string {
	if d.IsMemory() {
		d.Path = ":memory:"
	}
	return "file:" + d.Path + "?" + d.GetParamString()
}

// BuildFullDSN 生成带账号的完整DSN
func (d Sqlite) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if !d.IsMemory() && username != "" {
		dsn += "_auth_user=" + username + "&"
		password = url.QueryEscape(password)
		dsn += "_auth_pass=" + password + "&"
	}
	return dsn
}

// IsMemory 是否内存数据库
func (d Sqlite) IsMemory() bool {
	return d.Path == "" || strings.ToLower(d.Path) == ":memory:"
}

// GetCurrentDB 获得当前数据库名
func (Sqlite) GetCurrentDB(db *sql.DB) string {
	return ""
}

// FindTableInfos 查找表信息
func (d Sqlite) FindTableInfos(db *sql.DB) []*dbutil.TableSchema {
	query := `SELECT tbl_name, name FROM sqlite_master WHERE type = 'table'
AND tbl_name NOT LIKE 'sqlite_%' ORDER BY tbl_name`
	tables := dbutil.QueryTableInfos(db, query)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (Sqlite) FetchColumnInfos(db *sql.DB, table string) []*dbutil.ColumnInfo {
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

	var cols []*dbutil.ColumnInfo
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

func parseString(colStr string, pks map[string]bool) (*dbutil.ColumnInfo, error) {
	fields := splitColStr(colStr)
	col := &dbutil.ColumnInfo{Nullable: true}
	for idx, field := range fields {
		if idx == 0 {
			col.Name = strings.Trim(strings.TrimSpace(field), "`[]'\"")
			continue
		} else if idx == 1 {
			col.DataType = field
			continue
		}
		if _, ok := pks[col.Name]; ok {
			col.Index = sql.Null[string]{Valid: true, V: "pk"}
		}
		switch field {
		case "PRIMARY":
			col.Index = sql.Null[string]{Valid: true, V: "pk"}
		case "AUTOINCREMENT":
			col.Extra = sql.Null[string]{Valid: true, V: "auto_incr"}
		case "NULL":
			if fields[idx-1] == "NOT" {
				col.Nullable = false
			} else {
				col.Nullable = true
			}
		case "DEFAULT":
			col.Default = sql.Null[string]{Valid: true, V: fields[idx+1]}
		}
	}
	return col, nil
}

// splitColStr splits a sqlite col strings as fields
func splitColStr(colStr string) []string {
	results := make([]string, 0, 10)
	var lastIdx int
	var hasC, hasQuote bool
	for i, c := range colStr {
		if unicode.IsSpace(c) && !hasQuote {
			if hasC {
				results = append(results, colStr[lastIdx:i])
				hasC = false
			}
		} else {
			if c == '\'' {
				hasQuote = !hasQuote
			}
			if !hasC {
				lastIdx = i
			}
			hasC = true
			if i == len(colStr)-1 {
				results = append(results, colStr[lastIdx:i+1])
			}
		}
	}
	return results
}
