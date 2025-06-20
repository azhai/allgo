package tests_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/azhai/allgo/config"
	"github.com/azhai/allgo/dbutil"
	"github.com/azhai/allgo/services/db"
	_ "github.com/codenotary/immudb/pkg/stdlib"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	// _ "github.com/mattn/go-sqlite3"
)

var (
	logFile string
	env     *config.Environ
	dbSrc   *dbutil.DBServ
	dbDest  *dbutil.DBServ

	models = map[string]dbutil.ModelMigrator{
		"t_wall_daily": &db.WallDaily{},
		"t_wall_image": &db.WallImage{},
		"t_wall_note":  &db.WallNote{},
	}
	createTables = map[string]string{
		"t_wall_daily": `CREATE TABLE IF NOT EXISTS t_wall_daily (
  id integer NOT NULL,
  guid varchar[10] NOT NULL,
  bing_date timestamp,
  bing_sku varchar[100] NOT NULL,
  title varchar[255] NOT NULL,
  headline varchar[255] NOT NULL,
  color varchar[15] NOT NULL,
  max_dpi varchar[15] NOT NULL,
  PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS ON t_wall_daily (bing_date);
CREATE INDEX IF NOT EXISTS ON t_wall_daily (title);
`,
		"t_wall_image": `CREATE TABLE IF NOT EXISTS t_wall_image (
  id integer NOT NULL,
  daily_id integer NOT NULL,
  file_name varchar[100] NOT NULL,
  img_md5 varchar[32] NOT NULL,
  img_size integer NOT NULL,
  img_offset integer NOT NULL,
  img_width integer NOT NULL,
  img_height integer NOT NULL,
  PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS ON t_wall_image (daily_id);
CREATE INDEX IF NOT EXISTS ON t_wall_image (img_md5);
CREATE INDEX IF NOT EXISTS ON t_wall_image (img_size);
`,
		"t_wall_note": `CREATE TABLE IF NOT EXISTS t_wall_note (
  id integer NOT NULL AUTO_INCREMENT,
  daily_id integer NOT NULL,
  note_type varchar[50] NOT NULL,
  note_chinese varchar,
  note_english varchar,
  PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS ON t_wall_note (daily_id);
`,
	}
)

func init() {
	env = config.NewWithFile("../.env.testing")
	logFile = env.Get("DATABASE_LOG")
	dbSrc = dbutil.FromDialect("pgsql", env.Get("PGSQL_URL"))
	err := dbSrc.SetDB(sql.Open(dbSrc.Type, dbSrc.DSN))
	dbDest = dbutil.FromDialect("immudb", env.Get("IMMUDB_URL"))
	err = dbDest.SetDB(sql.Open(dbDest.Type, dbDest.DSN))
	if err == nil && logFile != "" {
		dbSrc.WithLogger(logFile)
		// dbDest.WithLogger(logFile)
	}
}

// go test -run=Migrate_Table
func Test11_Migrate_Table(t *testing.T) {
	tableInfos := dbSrc.LoadDialect().FindTableInfos(dbSrc.DB)
	for _, table := range tableInfos {
		fmt.Println(table.Name, table.Comment.V)

		// for _, col := range table.Columns {
		// 	fmt.Printf("%s %s(%d)\n", col.Name, col.ColType, col.Length.V)
		// }

		if query, ok := createTables[table.Name]; ok {
			_, err := dbDest.Exec(query)
			if err != nil {
				panic(err)
			}
		}

	}
}

// go test -run=Migrate_Data
func Test12_Migrate_Data(t *testing.T) {
	var limit = 500
	tableInfos := dbSrc.LoadDialect().FindTableInfos(dbSrc.DB)
	for _, table := range tableInfos {
		fmt.Println(table.Name, table.Comment.V)
		if model, ok := models[table.Name]; ok {
			switch model := model.(type) {
			case *db.WallDaily:
				dbutil.MigrateTableData(dbSrc, dbDest, model, limit)
			case *db.WallImage:
				dbutil.MigrateTableData(dbSrc, dbDest, model, limit)
				// case *db.WallNote:
				// 	dbutil.MigrateTableData(dbSrc, dbDest, model, limit)
			}
		}
	}
}
