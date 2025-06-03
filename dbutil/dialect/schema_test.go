package dialect_test

import (
	"fmt"
	"testing"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/dbutil/dialect"
)

var (
	env *config.Environ
	db  *dbutil.DBServ
	dia dialect.Dialect
)

func init() {
	env = config.NewWithFile("../../.env.testing")
	dbType := env.GetStr("DATABASE_TYPE", "postgres")
	dsn := env.Get("DATABASE_URL")
	db, _ = dbutil.New(dbType, dsn)
	logFile := env.Get("DATABASE_LOG_FILE")
	db.WithLogger("info", logFile)
	dia = dialect.CreateDialectByName(dbType)
}

// go test -run=Postgres
func Test11_Postgres(t *testing.T) {
	infos := dia.FindTableInfos(db)
	for _, info := range infos {
		fmt.Println(info.Name, *info.Comment)
		for _, col := range info.Columns {
			fmt.Printf("%+v\n", col)
		}
	}
}
