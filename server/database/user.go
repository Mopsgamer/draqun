package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

// Create new DB record.
func (db Database) UserCreate(user model_database.User) *uint64 {
	return Insert(db, "app_users", user)
}

// Change the existing DB record.
func (db Database) UserUpdate(user model_database.User) bool {
	return Update(db, "app_users", user, goqu.Ex(exp.Ex{"id": user.Id}))
}

// Delete the existing DB record.
func (db Database) UserDelete(userId uint64) bool {
	return Delete(db, "app_users", goqu.Ex(exp.Ex{"id": userId}))
}

// Get the user by his email.
func (db Database) UserByEmail(email string) *model_database.User {
	return First[model_database.User](db, "app_users", goqu.Ex{"email": email})
}

// Get the user by his identificator.
func (db Database) UserById(userId uint64) *model_database.User {
	return First[model_database.User](db, "app_users", goqu.Ex{"id": userId})
}

// Get the user by his username.
func (db Database) UserByUsername(username string) *model_database.User {
	return First[model_database.User](db, "app_users", goqu.Ex{"username": username})
}

func (db Database) UserOwnGroupList(userId uint64) []model_database.Group {
	groupList := &[]model_database.Group{}
	query := `SELECT
		app_groups.*
	FROM app_groups
	LEFT JOIN app_group_members ON app_groups.id = app_group_members.group_id
	WHERE (user_id = ? AND is_owner = 1)`
	err := db.Sqlx.Select(groupList, query, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return *groupList
		}
		log.Error(err)
		return *groupList
	}
	return *groupList
}

func (db Database) UserGroupList(userId uint64) []model_database.Group {
	groupList := &[]model_database.Group{}
	query := `SELECT
		app_groups.*
	FROM app_groups
	LEFT JOIN app_group_members ON app_groups.id = app_group_members.group_id
	WHERE user_id = ?`
	err := db.Sqlx.Select(groupList, query, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return *groupList
		}
		log.Error(err)
		return *groupList
	}
	return *groupList
}
