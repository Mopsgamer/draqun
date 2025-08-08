package model

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/doug-martin/goqu/v9/exp"
	"golang.org/x/exp/constraints"
)

func First[T any](db *DB, table string, ex goqu.Ex, item *T) (err error) {
	sql, args, err := db.Goqu.From(table).Select(item).Where(ex).ToSQL()
	if err != nil {
		return err
	}

	err = db.Sqlx.DB.QueryRow(sql, args...).Scan(item)
	return
}

func Last[T any](db *DB, table string, ex goqu.Ex, key exp.IdentifierExpression, item *T) (err error) {
	sql, args, err := db.Goqu.From(table).Select(item).Where(ex).Order(key.Desc()).Limit(1).ToSQL()
	if err != nil {
		return
	}

	err = db.Sqlx.DB.QueryRow(sql, args...).Scan(item)
	return
}

func insert[T any](db *DB, table string, item *T) (result sql.Result, err error) {
	sql, args, err := db.Goqu.Insert(table).Rows(item).ToSQL()
	if err != nil {
		return
	}

	result, err = db.Sqlx.Exec(sql, args)
	if err != nil {
		return
	}

	return
}

func Insert0[T any](db *DB, table string, item *T) (err error) {
	_, err = insert(db, table, item)
	return
}

func InsertId[T any, Id constraints.Integer](db *DB, table string, item *T, itemId *Id) (err error) {
	result, err := insert(db, table, item)
	if err != nil {
		return
	}

	newId, err := result.LastInsertId()
	if err != nil {
		return
	}

	*itemId = Id(newId)
	return
}

func Update[T any](db *DB, table string, item T, ex goqu.Ex) (err error) {
	sql, args, err := db.Goqu.Update(table).Set(item).Where(ex).ToSQL()
	if err != nil {
		return
	}

	_, err = db.Sqlx.Exec(sql, args)
	if err != nil {
		return
	}

	return
}

func Delete[T string | []any](db *DB, table T, ex goqu.Ex) bool {
	sql, args, err := db.Goqu.From(table).Delete().Where(ex).ToSQL()
	if err != nil {
		return false
	}

	_, err = db.Sqlx.Exec(sql, args...)
	return err == nil
}
