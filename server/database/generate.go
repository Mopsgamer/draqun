package database

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/gofiber/fiber/v3/log"
)

func First[T any](db Database, table string, ex goqu.Ex) *T {
	item := new(T)
	query, params, err := db.Goqu.From(table).Prepared(true).Select(item).Where(ex).ToSQL()
	if err != nil {
		log.Error(err, query)
		return nil
	}

	err = db.Sqlx.Get(item, query, params...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		log.Error(err, query)
		return nil
	}
	return item
}

func Insert[T any](db Database, table string, item T) *uint64 {
	query, params, err := db.Goqu.Insert(table).Prepared(true).Rows(item).ToSQL()
	if err != nil {
		log.Error(err, query)
		return nil
	}

	r, err := db.Sqlx.Exec(query, params...)
	if err != nil {
		log.Error(err, query)
		return nil
	}

	newId, err := r.LastInsertId()
	if err != nil {
		log.Error(err, query)
		return nil
	}

	newIdUint := uint64(newId)
	return &newIdUint
}

func Update[T any](db Database, table string, item T, ex goqu.Ex) bool {
	query, params, err := db.Goqu.Update(table).Prepared(true).Set(item).Where(ex).ToSQL()
	if err != nil {
		log.Error(err, query)
		return false
	}

	_, err = db.Sqlx.Exec(query, params...)
	if err != nil {
		log.Error(err, query)
		return false
	}

	return true
}

func Delete(db Database, table string, ex goqu.Ex) bool {
	query, params, err := db.Goqu.From(table).Prepared(true).Delete().Where(ex).ToSQL()
	if err != nil {
		log.Error(err, query)
		return false
	}

	_, err = db.Sqlx.Exec(query, params...)
	if err != nil {
		log.Error(err, query)
		return false
	}

	return true
}
