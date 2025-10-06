package model

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"golang.org/x/exp/constraints"
)

func First[T any](table string, ex goqu.Ex, item *T) (err error) {
	sql, args, err := Goqu.From(table).Select(item).Where(ex).Prepared(true).ToSQL()
	if err != nil {
		return err
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(item)
	return
}

func Last[T any](table string, ex goqu.Ex, key exp.IdentifierExpression, item *T) (err error) {
	sql, args, err := Goqu.From(table).Select(item).Where(ex).Order(key.Desc()).Limit(1).Prepared(true).ToSQL()
	if err != nil {
		return
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(item)
	return
}

func insert[T any](table string, item *T) (result sql.Result, err error) {
	sql, args, err := Goqu.Insert(table).Rows(item).Prepared(true).ToSQL()
	if err != nil {
		return
	}

	result, err = Sqlx.Exec(sql, args...)
	if err != nil {
		return
	}

	return
}

func Insert[T any](table string, item *T) (err error) {
	_, err = insert(table, item)
	return
}

func InsertId[T any, Id constraints.Integer](table string, item *T, itemId *Id) (err error) {
	result, err := insert(table, item)
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

func Update[T any](table string, item T, ex goqu.Ex) (err error) {
	sql, args, err := Goqu.Update(table).Set(item).Where(ex).Prepared(true).ToSQL()
	if err != nil {
		return
	}

	_, err = Sqlx.Exec(sql, args...)
	if err != nil {
		return
	}

	return
}

func Delete[T string | []any](table T, ex goqu.Ex) (err error) {
	sql, args, err := Goqu.From(table).Delete().Where(ex).Prepared(true).ToSQL()
	if err != nil {
		return
	}

	_, err = Sqlx.Exec(sql, args...)
	return
}
