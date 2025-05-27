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
	dbType := env.GetStr("DATABASE_TYPE", "postgres")
	dbDSN := env.Get("DATABASE_URL")
	// fmt.Println(dbType, dbDSN)
	db, err := sql.Open(dbType, dbDSN)
	if err == nil && db != nil {
		logLevel := env.GetStr("LOG_LEVEL", "info")
		logFile := env.Get("DATABASE_LOG_FILE")
		// fmt.Println(logLevel, logFile)
		dbServ = &dbutil.DBServ{DB: db, DSN: dbDSN}
		dbServ.WithLogger(logFile, logLevel)
		ctx := context.Background()
		err = dbServ.PingContext(ctx)
	}
	return err
}

// CloseService 关闭服务
func CloseService() {
	if dbServ != nil {
		_ = dbServ.Close()
		dbServ = nil
	}
}
