package dbutil

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var NotPtrList = errors.New("dest must be a pointer to a slice")

type Model interface {
	// TableName 返回表名
	TableName() string
}

type ModelComment interface {
	Model
	// TableComment 返回表备注
	TableComment() string
}

type ModelChanger interface {
	Model
	// UniqFields 可作为更新条件的字段与它的值
	UniqFields() ([]string, []any)
	// SetId 设置主键值
	SetId(id int64, err error) error
	// RowValues 插入一行所需数据
	RowValues() []any
	// InsertSQL 插入一行的SQL语句
	InsertSQL() string
	// UpsertSQL 插入或更新一行的SQL语句
	UpsertSQL() string
}

func (s *DBServ) ExecUpdate(table, where string, wargs []sql.NamedArg,
	changes map[string]any) (int, error) {
	var nargs []sql.NamedArg
	query := "UPDATE " + table + " SET "
	for k, v := range changes {
		query += fmt.Sprintf("%s = @%s, ", k, k)
		nargs = append(nargs, sql.Named(k, v))
	}
	query = strings.TrimSuffix(query, ", ") + " WHERE " + where
	if len(wargs) > 0 {
		nargs = append(nargs, wargs...)
	}
	res, err := s.NamedExec(nil, query, nargs)
	var num int64
	if err == nil {
		num, err = res.RowsAffected()
	}
	return int(num), err
}

func (s *DBServ) UpdateRow(row ModelChanger, changes map[string]any) (int, error) {
	table, where := row.TableName(), ""
	cols, values := row.UniqFields()
	// WHERE参数可能和SET参数有相同字段
	var wargs []sql.NamedArg
	for i, k := range cols {
		kk, val := fmt.Sprintf("_%s", k), values[i]
		where += fmt.Sprintf("%s = @%s AND ", k, kk)
		wargs = append(wargs, sql.Named(kk, val))
	}
	where = strings.TrimSuffix(where, " AND ")
	return s.ExecUpdate(table, where, wargs, changes)
}

func (s *DBServ) execInsert(row ModelChanger, stmt *sql.Stmt, query string) (bool, error) {
	var (
		err error
		res sql.Result
	)
	args := row.RowValues()
	if stmt != nil {
		res, err = stmt.Exec(args...)
		// if res != nil {
		// 	err = row.SetId(res.LastInsertId())
		// }
	} else if len(query) > 0 {
		res, err = s.Exec(query, args...)
	}
	if err != nil || res == nil {
		return false, err
	}
	return true, err
}

func (s *DBServ) execInsertId(row ModelChanger, stmt *sql.Stmt, query string) (int64, error) {
	var (
		err    error
		lastId int64
	)
	args := row.RowValues()
	if stmt != nil {
		err = stmt.QueryRow(args...).Scan(&lastId)
	} else if len(query) > 0 {
		err = s.QueryRow(query, args...).Scan(&lastId)
	}
	err = row.SetId(lastId, err)
	return lastId, err
}

func (s *DBServ) UpsertRow(row ModelChanger) (bool, error) {
	return s.execInsert(row, nil, row.UpsertSQL())
}

func (s *DBServ) InsertRow(row ModelChanger) (bool, error) {
	query := row.InsertSQL()
	if !strings.Contains(query, " RETURNING ") {
		return s.execInsert(row, nil, query)
	}
	_, err := s.execInsertId(row, nil, query)
	return err == nil, err
}

func InsertBatch[T ModelChanger](db *DBServ, rows []T) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	query := rows[0].InsertSQL()
	stmt, err := db.Prepare(query)
	if err != nil {
		return 0, err
	}
	num, withId := 0, strings.Contains(query, " RETURNING ")
	for _, row := range rows {
		if !withId {
			_, err = db.execInsert(row, stmt, query)
		} else {
			_, err = db.execInsertId(row, stmt, query)
		}
		if err != nil {
			break
		}
		num++
	}
	err = stmt.Close()
	return num, err
}

// ScanSource 扫描源，即sql.Rows或sql.Row
type ScanSource interface {
	Scan(dest ...any) error
	Err() error
}

// ModelLoader 可分解原始数据的Model
// 用于从sql.Rows中读取数据
type ModelLoader interface {
	// ScanFrom 从src中读取数据写入当前对象
	ScanFrom(src ScanSource, err error) error
}

// ModelForeignLoader 外键扫描Model
type ModelForeignLoader interface {
	ModelLoader
	// ForeignIndex 返回外键的值
	ForeignIndex() any
}

// ModelSecondaryLoader 外键扫描Model
type ModelSecondaryLoader interface {
	ModelForeignLoader
	// SecondaryKey 返回次要字段的值
	SecondaryKey() string
}

// ScanToList 扫描结果集到列表
// dest必须是一个指向切片的指针
func ScanToList[T ModelLoader](dest *[]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem().Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		*dest = append(*dest, elem)
	}
	return rs.Err()
}

// ScanToUnique  扫描结果集到一对一外键Map
func ScanToUnique[K comparable, T ModelForeignLoader](dest map[K]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		idx := elem.ForeignIndex().(K)
		dest[idx] = elem
	}
	return rs.Err()
}

// ScanToIndex 扫描结果集到一对多外键Map
func ScanToIndex[K comparable, T ModelForeignLoader](dest map[K][]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem().Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		idx := elem.ForeignIndex().(K)
		dest[idx] = append(dest[idx], elem)
	}
	return rs.Err()
}

// ScanToSecondary 扫描结果集到双层外键Map
func ScanToSecondary[K comparable, T ModelSecondaryLoader](dest map[K]map[string]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem().Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		idx := elem.ForeignIndex().(K)
		if _, ok := dest[idx]; !ok {
			dest[idx] = make(map[string]T)
		}
		key := elem.SecondaryKey()
		dest[idx][key] = elem
	}
	return rs.Err()
}

// ScanToStructs 扫描结果集到Struct
func ScanToStructs[T any](dest *[]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem().Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	var err error
	for rs.Next() {
		elem := reflect.New(dt.Elem()).Interface().(T)
		vt := reflect.ValueOf(elem)
		if vt.Kind() != reflect.Struct {
			err = rs.Scan(&elem)
		} else {
			num := vt.NumField()
			columns := make([]any, num)
			for i := 0; i < num; i++ {
				columns[i] = vt.Field(i).Addr().Interface()
			}
			err = rs.Scan(columns...)
		}
		if err != nil {
			return err
		}
		*dest = append(*dest, elem)
	}
	return rs.Err()
}

// ScanToMap 扫描结果集到Map
// dest必须是一个指向Map的指针
func ScanToMap[T any](dest map[string]T, rs *sql.Rows) error {
	defer rs.Close()
	for rs.Next() {
		var (
			key   string
			value T
		)
		err := rs.Scan(&key, &value)
		if err != nil {
			return err
		}
		dest[key] = value
	}
	return rs.Err()
}
