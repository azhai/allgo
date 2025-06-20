package dbutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/azhai/allgo/dbutil/dialect"
	"github.com/azhai/allgo/logutil"
	"github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zapadapter"
)

type DBServ struct {
	DSN          string
	Name         string // 当前数据库
	Type         string // 数据库类型
	dollarHolder bool
	Dialect      Dialect
	*sql.DB
}

// FromDSN 创建一个不带sql.DB的对象
func FromDSN(dbType, dsn string) *DBServ {
	if dbType == "" {
		dbType = ParseScheme(dsn)
	}
	return &DBServ{DSN: dsn, Type: dbType}
}

// FromDialect 创建一个不带sql.DB的对象，使用dialect补充连接参数
func FromDialect(dbType, dsn string) *DBServ {
	u, err := url.Parse(dsn)
	if err != nil || u == nil {
		panic(err)
	}
	if dbType == "" {
		dbType = u.Scheme
	}

	obj := &DBServ{DSN: dsn, Type: dbType}
	if dia := obj.LoadDialect(); dia != nil {
		_ = dia.ParseUrlQuery(u.RawQuery)
		size := len(obj.DSN) - len(u.RawQuery)
		head := strings.TrimRight(obj.DSN[:size], "?& ")
		obj.DSN = head + "?" + dia.GetParamString()
	}
	return obj
}

// LoadDialect 加载数据库类型
// 需要在执行的文件开头注册已知的dialect
// import _ "github.com/azhai/allgo/dbutil/dialect"
func (s *DBServ) LoadDialect() Dialect {
	if s.Dialect != nil {
		return s.Dialect
	}
	if dia, ok := dialects[s.Type]; ok {
		s.Dialect, s.Type = dia, dia.TypeName()
		return s.Dialect
	}
	panic(fmt.Errorf("unsupported database type: %s", s.Type))
}

// SetDB 替换数据库对象
func (s *DBServ) SetDB(db *sql.DB, err error) error {
	if err != nil || db == nil {
		return err
	}
	s.DB = db
	ctx := context.Background()
	return s.PingContext(ctx)
}

// WithLogger 记录SQL到日志
func (s *DBServ) WithLogger(filename string) {
	logger := logutil.NewLoggerURL(filename)
	loggerAdapter := zapadapter.New(logger.Desugar())
	s.DB = sqldblogger.OpenDriver(s.DSN, s.DB.Driver(), loggerAdapter)
}

// AllowDollarHolder SQL中是否可用$1、$2等占位符
func (s *DBServ) AllowDollarHolder() bool {
	if s.Dialect == nil {
		dia := s.LoadDialect()
		s.dollarHolder = dia.IsSupport(dialect.FeatDollarHolder)
	}
	return s.dollarHolder
}

func (s *DBServ) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.ExecContext(ctx, query, args...)
}

func (s *DBServ) Exec(query string, args ...any) (sql.Result, error) {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.Exec(query, args...)
}

func (s *DBServ) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.PrepareContext(ctx, query)
}

func (s *DBServ) Prepare(query string) (*sql.Stmt, error) {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.Prepare(query)
}

func (s *DBServ) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.QueryContext(ctx, query, args...)
}

func (s *DBServ) Query(query string, args ...any) (*sql.Rows, error) {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.Query(query, args...)
}

func (s *DBServ) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.QueryRowContext(ctx, query, args...)
}

func (s *DBServ) QueryRow(query string, args ...any) *sql.Row {
	if !s.AllowDollarHolder() {
		query = QuestionMarkHolders(query)
	}
	return s.DB.QueryRow(query, args...)
}

func (s *DBServ) FlattenExec(ctx context.Context,
	query string, args ...any) (sql.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	query, args = FlattenPlaceHolders(query, args)
	return s.ExecContext(ctx, query, args...)
}

func (s *DBServ) FlattenQuery(ctx context.Context,
	query string, args ...any) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	query, args = FlattenPlaceHolders(query, args)
	return s.QueryContext(ctx, query, args...)
}

func (s *DBServ) NamedExec(ctx context.Context,
	query string, nargs []sql.NamedArg) (sql.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var args []any
	query, args = RegularNamedPlaceHolders(query, nargs)
	return s.ExecContext(ctx, query, args...)
}

func (s *DBServ) NamedQuery(ctx context.Context,
	query string, nargs []sql.NamedArg) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var args []any
	query, args = RegularNamedPlaceHolders(query, nargs)
	return s.QueryContext(ctx, query, args...)
}

// QuestionMarkHolders 将SQL中的顺序占位符替换为问号占位符
func QuestionMarkHolders(query string) string {
	re := regexp.MustCompile(`\$\d+`)
	return re.ReplaceAllString(query, "?")
}

// FlattenPlaceHolders 将SQL中对应slice参数的顺序占位符进行扩充
func FlattenPlaceHolders(query string, args []any) (string, []any) {
	var (
		pairs  []string
		values []any
	)
	for i, arg := range args {
		old := fmt.Sprintf("$%d", i+1)
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Slice {
			values = append(values, arg)
			holder := "$" + strconv.Itoa(len(values))
			if old != holder {
				pairs = append(pairs, holder, old)
			}
			continue
		}
		var holderList []string
		for j := 0; j < rv.Len(); j++ {
			values = append(values, rv.Index(j).Interface())
			holder := "$" + strconv.Itoa(len(values))
			holderList = append(holderList, holder)
		}
		pairs = append(pairs, strings.Join(holderList, ", "), old)
	}
	slices.Reverse(pairs)
	replacer := strings.NewReplacer(pairs...)
	return replacer.Replace(query), values
}

// RegularNamedPlaceHolders 将SQL中的命名占位符替换为pq driver库的顺序占位符
func RegularNamedPlaceHolders(query string, nargs []sql.NamedArg) (string, []any) {
	// 倒序排列防止短匹配优先 overlapping matches
	slices.SortFunc(nargs, func(a, b sql.NamedArg) int {
		return 0 - strings.Compare(a.Name, b.Name)
	})
	var (
		pairs  []string
		values []any
	)
	for _, arg := range nargs {
		values = append(values, arg.Value)
		holder := "$" + strconv.Itoa(len(values))
		pairs = append(pairs, "@"+arg.Name, holder)
	}
	replacer := strings.NewReplacer(pairs...)
	return replacer.Replace(query), values
}
