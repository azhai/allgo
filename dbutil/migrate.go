package dbutil

import (
	"fmt"
)

type ModelMigrator interface {
	ModelLoader
	ModelChanger
}

// MigrateTableData 迁移表数据
// 从dbSrc中读取表table的数据，写入dbDest中
func MigrateTableData[T ModelMigrator](dbSrc, dbDest *DBServ, model T, limit int) {
	pk := model.PrimaryKey()
	query := "SELECT * FROM %s WHERE %s > $1 ORDER BY %s LIMIT $2"
	query = fmt.Sprintf(query, model.TableName(), pk, pk)

	var lastId int64
	for {
		var objs []T
		rows, err := dbSrc.Query(query, lastId, limit)
		if err == nil {
			err = ScanToList(&objs, rows)
		}
		if err != nil {
			panic(err)
		}
		if len(objs) == 0 {
			break
		}
		lastId = objs[len(objs)-1].GetId()
		_, err = InsertBatch(dbDest, objs)
		if err != nil {
			panic(err)
		}
	}
}
