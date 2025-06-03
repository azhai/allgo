package dialect

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/azhai/allgo/dbutil"
)

// Sqlite SQLite3数据库
type Sqlite struct {
	Path string `hcl:"path,optional" json:"path"`
}

// IsRelationalDB 是否关系数据库
func (Sqlite) IsRelationalDB() bool {
	return true
}

// Name 驱动名
func (Sqlite) Name() string {
	return "sqlite3"
}

// ImporterPath 驱动支持库
func (Sqlite) ImporterPath() string {
	return "github.com/mattn/go-sqlite3"
}

// QuoteIdent 字段或表名脱敏
func (Sqlite) QuoteIdent(ident string) string {
	return WrapWith(ident, "`", "`")
}

// BuildDSN 生成DSN连接串
func (d Sqlite) BuildDSN() string {
	if d.IsMemory() {
		d.Path = ":memory:"
	}
	return "file:" + d.Path + "?cache=shared&"
}

// BuildFullDSN 生成带账号的完整DSN
func (d Sqlite) BuildFullDSN(username, password string) string {
	dsn := d.BuildDSN()
	if !d.IsMemory() && username != "" {
		dsn += "_auth_user=" + username + "&"
		dsn += "_auth_pass=" + Escape(password) + "&"
	}
	return dsn
}

// IsMemory 是否内存数据库
func (d Sqlite) IsMemory() bool {
	return d.Path == "" || strings.ToLower(d.Path) == ":memory:"
}

// GetCurrentDB 获得当前数据库名
func (Sqlite) GetCurrentDB(db *dbutil.DBServ) string {
	return ""
}

// FindTableInfos 查找表信息
func (d Sqlite) FindTableInfos(db *dbutil.DBServ) []*TableSchema {
	query := `SELECT tbl_name, name FROM sqlite_master WHERE type='table'`
	tables := QueryTableInfos(db, query, db.Name)
	for i, table := range tables {
		table.Columns = d.FetchColumnInfos(db, table.Name)
		tables[i] = table
	}
	return tables
}

// FetchColumnInfos 查找字段信息
func (Sqlite) FetchColumnInfos(db *dbutil.DBServ, table string) []*ColumnInfo {
	query := `SELECT sql FROM sqlite_master WHERE type='table' AND tbl_name = ?`
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

func parseString(colStr string, pks map[string]bool) (*ColumnInfo, error) {
	fields := splitColStr(colStr)
	col := &ColumnInfo{Nullable: true}
	pk, auto_incr := "pk", "auto_incr"
	for idx, field := range fields {
		if idx == 0 {
			col.Name = strings.Trim(strings.TrimSpace(field), "`[]'\"")
			continue
		} else if idx == 1 {
			col.DataType = field
			continue
		}
		if _, ok := pks[col.Name]; ok {
			col.Index = &pk
		}
		switch field {
		case "PRIMARY":
			col.Index = &pk
		case "AUTOINCREMENT":
			col.Extra = &auto_incr
		case "NULL":
			if fields[idx-1] == "NOT" {
				col.Nullable = false
			} else {
				col.Nullable = true
			}
		case "DEFAULT":
			col.Default = &fields[idx+1]
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
