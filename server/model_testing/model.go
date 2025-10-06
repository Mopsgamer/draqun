package modeltesting

import (
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/jmoiron/sqlx"
)

type Query func(tx *sqlx.Tx)

func NewTempDB(use ...Query) error {
	tx, err := model.Sqlx.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, u := range use {
		u(tx)
	}
	return nil
}
