package database

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

func First[T any](db *goqu.Database, table string, ex goqu.Ex, item *T) {
	found, err := db.From(table).Select(item).Where(ex).ScanStruct(item)
	if !found {
		log.Error(err)
	}
}

func Last[T any](db *goqu.Database, table string, ex goqu.Ex, key exp.IdentifierExpression, item *T) {
	found, err := db.From(table).Select(item).Where(ex).Order(key.Desc()).Limit(1).ScanStruct(item)
	if !found {
		log.Error(err)
	}
	if err != nil {
		log.Error(err)
	}
}

func Insert[T any](db *goqu.Database, table string, item *T) *uint64 {
	result, err := db.Insert(table).Rows(item).Executor().Exec()
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

func Update[T any](db *goqu.Database, table string, item T, ex goqu.Ex) bool {
	_, err := db.Update(table).Set(item).Where(ex).Executor().Exec()
	isErr := err != nil
	if isErr {
		log.Error(err)
	}

	return isErr
}

func Delete[T string | []any](db *goqu.Database, table T, ex goqu.Ex) bool {
	_, err := db.From(table).Delete().Where(ex).Executor().Exec()
	isErr := err != nil
	if isErr {
		log.Error(err)
	}

	return isErr
}
