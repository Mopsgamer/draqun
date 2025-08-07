package model

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

func First[T any](db *DB, table string, ex goqu.Ex, item *T) {
	sql, args, err := db.Goqu.From(table).Select(item).Where(ex).ToSQL()
	if err != nil {
		log.Error(err)
		return
	}

	err = db.Sqlx.DB.QueryRow(sql, args...).Scan(item)
	if err != nil {
		log.Error(err)
		return
	}
}

func Last[T any](db *DB, table string, ex goqu.Ex, key exp.IdentifierExpression, item *T) {
	sql, args, err := db.Goqu.From(table).Select(item).Where(ex).Order(key.Desc()).Limit(1).ToSQL()
	if err != nil {
		log.Error(err)
		return
	}

	err = db.Sqlx.DB.QueryRow(sql, args...).Scan(item)
	if err != nil {
		log.Error(err)
		return
	}
}

func Insert[T any](db *DB, table string, item *T) uint64 {
	sql, args, err := db.Goqu.Insert(table).Rows(item).ToSQL()
	if err != nil {
		log.Error(err)
		return 0
	}

	result, err := db.Sqlx.Exec(sql, args)
	if err != nil {
		log.Error(err)
		return 0
	}

	newId, err := result.LastInsertId()
	if err != nil {
		log.Error(err)
		return 0
	}

	newIdUint := uint64(newId)
	return newIdUint
}

func Update[T any](db *DB, table string, item T, ex goqu.Ex) bool {
	sql, args, err := db.Goqu.Update(table).Set(item).Where(ex).ToSQL()
	if err != nil {
		log.Error(err)
		return false
	}

	_, err = db.Sqlx.Exec(sql, args)
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}

func Delete[T string | []any](db *DB, table T, ex goqu.Ex) bool {
	sql, args, err := db.Goqu.From(table).Delete().Where(ex).ToSQL()
	if err != nil {
		log.Error(err)
		return false
	}

	_, err = db.Sqlx.Exec(sql, args...)
	if err != nil {
		log.Error(err)
		return false
	}

	return true
}
