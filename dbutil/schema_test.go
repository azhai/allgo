package dbutil_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	env    *config.Environ
	dbServ *dbutil.DBServ
	dia    dbutil.Dialect
)

func init() {
	env = config.NewWithFile("../.env.testing")
	dsn, dbType := env.Get("DATABASE_URL"), env.Get("DATABASE_TYPE")
	logFile := env.Get("DATABASE_LOG")
	if err := newDB(dsn, dbType, logFile); err != nil {
		panic(err)
	}
	dia = dbutil.CreateDialectByName(dbServ.Type)
}

func newDB(dsn, dbType, logFile string) error {
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
	if logFile != "" {
		dbServ.WithLogger(logFile)
	}
	return nil
}

// go test -run=Columns
func Test11_Columns(t *testing.T) {
	infos := dia.FindTableInfos(dbServ)
	for _, info := range infos {
		fmt.Println(info.Name, info.Comment.V)
		for _, col := range info.Columns {
			fmt.Printf("%+v\n", col)
		}
	}
}
