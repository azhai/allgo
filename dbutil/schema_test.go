package dbutil_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	_ "github.com/azhai/allgo/dbutil/dialect"
	_ "github.com/codenotary/immudb/pkg/stdlib"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	env    *config.Environ
	dbServ *dbutil.DBServ
)

func init() {
	env = config.NewWithFile("../.env.testing")
	dsn, dbType := env.Get("DATABASE_URL"), env.Get("DATABASE_TYPE")
	logFile := env.Get("DATABASE_LOG")
	if err := newDB(dsn, dbType, logFile); err != nil {
		panic(err)
	}
}

func newDB(dsn, dbType, logFile string) error {
	dbServ = dbutil.FromDialect(dsn, dbType)
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
