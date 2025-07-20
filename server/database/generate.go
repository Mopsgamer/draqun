package database

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/gofiber/fiber/v3/log"
)

func First[T any](db Database, table string, ex goqu.Ex) *T {
	item := new(T)

	found, err := db.Goqu.From(table).Prepared(true).Select(item).Where(ex).ScanStruct(item)

	if !found {
		log.Error(err)
		return nil
	}

	return item
}

func Insert[T any](db Database, table string, item T) *uint64 {
	result, err := db.Goqu.Insert(table).Prepared(true).Rows(item).Executor().Exec()
	if err != nil {
		log.Error(err)
		return nil
	}

	newId, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return nil
	}

	newIdUint := uint64(newId)
	return &newIdUint
}

func Update[T any](db Database, table string, item T, ex goqu.Ex) bool {
	_, err := db.Goqu.Update(table).Prepared(true).Set(item).Where(ex).Executor().Exec()
	isErr := err != nil
	if isErr {
		log.Error(err)
	}

	return isErr
}

func Delete(db Database, table string, ex goqu.Ex) bool {
	_, err := db.Goqu.From(table).Prepared(true).Delete().Where(ex).Executor().Exec()
	isErr := err != nil
	if isErr {
		log.Error(err)
	}

	return isErr
}
