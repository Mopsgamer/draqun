package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/controller/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

func (db Database) MemberById(groupId, userId uint64) *model_database.Member {
	return First[model_database.Member](db, "app_group_members", goqu.Ex{"group_id": groupId, "user_id": userId})
}

func (db Database) MemberCreate(member model_database.Member) bool {
	return Insert(db, "app_group_members", member) != nil
}

func (db Database) MemberDelete(userId, groupId uint64) bool {
	return Delete(db, "app_group_members", exp.Ex{"group_id": groupId, "user_id": userId})
}

func (db Database) MemberList(groupId uint64) []model_database.User {
	memberList := &[]model_database.User{}
	query := `SELECT app_users.*
	FROM app_group_members
	LEFT JOIN app_users ON app_users.id = app_group_members.user_id
	WHERE app_group_members.group_id = ?`
	err := db.Sqlx.Select(memberList, query, groupId)

	if err != nil {
		if err == sql.ErrNoRows {
			return *memberList
		}
		log.Error(err)
		return *memberList
	}
	return *memberList
}

func (db Database) MemberListPage(groupId uint64, page uint64, perPage uint64) []model_database.User {
	memberList := &[]model_database.User{}
	query := `SELECT app_users.* FROM app_group_members
	LEFT JOIN app_users ON app_users.id = app_group_members.user_id
	WHERE app_group_members.group_id = ?
	ORDER BY app_group_members.user_id ASC LIMIT ?, ?`
	from := (page - 1) * perPage
	to := page * perPage
	err := db.Sqlx.Select(memberList, query, groupId, from, to)
	if err != nil {
		if err == sql.ErrNoRows {
			return *memberList
		}
		log.Error(err)
		return *memberList
	}
	return *memberList
}
