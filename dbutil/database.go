package dbutil

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/azhai/allgo/logutil"
	_ "github.com/lib/pq"
	"github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zapadapter"
)

type DBServ struct {
	DSN string
	*sql.DB
}

func (s *DBServ) WithLogger(filename, level string) {
	logger := logutil.NewLoggerURL(level, filename)
	loggerAdapter := zapadapter.New(logger.Desugar())
	s.DB = sqldblogger.OpenDriver(s.DSN, s.DB.Driver(), loggerAdapter)
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
