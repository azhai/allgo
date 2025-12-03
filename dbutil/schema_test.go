package dbutil_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var (
	env    *config.Environ
	dbServ *dbutil.DBServ
)

func init() {
	env = config.NewWithFile("../.env.testing")
	dbType, dsn := env.Get("DATABASE_TYPE"), env.Get("DATABASE_URL")
	logFile := env.Get("DATABASE_LOG")
	if err := newDB(dbType, dsn, logFile); err != nil {
		panic(err)
	}
}

func newDB(dbType, dsn, logFile string) error {
	dbServ = dbutil.FromDialect(dbType, dsn)
	err := dbServ.SetDB(sql.Open(dbServ.Type, dbServ.DSN))
	if err != nil || dbServ.DB == nil {
		return err
	}
	if logFile != "" {
		dbServ.WithLogger(logFile)
	}
	return nil
}

// go test -run=Columns
func Test11_Columns(t *testing.T) {
	dia := dbServ.LoadDialect()
	infos := dia.FindTableInfos(dbServ.DB)
	for _, info := range infos {
		fmt.Println(info.Name, info.Comment.V)
		for _, col := range info.Columns {
			fmt.Printf("%+v\n", col)
		}
	}
}
