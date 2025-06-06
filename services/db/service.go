package db

import (
	"database/sql"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	_ "github.com/azhai/allgo/dbutil/dialect"
	// _ "github.com/codenotary/immudb/pkg/stdlib"
	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	// _ "github.com/mattn/go-sqlite3"
)

func NewNullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: true}
}

var dbServ *dbutil.DBServ

// DB 获取数据库服务
func DB() *dbutil.DBServ {
	if dbServ == nil {
		err := OpenService(config.New())
		if err != nil {
			panic(err)
		}
	}
	return dbServ
}

// OpenService 初始化服务
func OpenService(env *config.Environ) error {
	dbType := env.GetStr("DATABASE_TYPE")
	dsn := env.Get("DATABASE_URL")
	dbServ = dbutil.FromDialect(dsn, dbType)
	err := dbServ.SetDB(sql.Open(dbServ.Type, dbServ.DSN))
	if err != nil || dbServ.DB == nil {
		return err
	}
	if logFile := env.Get("DATABASE_LOG"); logFile != "" {
		dbServ.WithLogger(logFile)
	}
	return nil
}

// CloseService 关闭服务
func CloseService() {
	if dbServ != nil {
		_ = dbServ.Close()
		dbServ = nil
	}
}
