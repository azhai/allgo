package db

import (
	"context"
	"database/sql"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	_ "github.com/lib/pq"
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
	dsn := env.Get("DATABASE_URL")
	dbType := env.GetStr("DATABASE_TYPE")
	if dbType == "" {
		dbType = dbutil.ParseSchema(dsn)
	}

	db, err := sql.Open(dbType, dsn)
	if err != nil || db == nil {
		return err
	}
	ctx := context.Background()
	if err = db.PingContext(ctx); err != nil {
		return err
	}

	dbServ = &dbutil.DBServ{DB: db, DSN: dsn, Type: dbType}
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
