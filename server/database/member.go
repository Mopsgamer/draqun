package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/gofiber/fiber/v3/log"
)

func (db Database) MemberById(groupId, userId uint64) *model_database.Member {
	return First[model_database.Member](db, TableMembers, goqu.Ex{"group_id": groupId, "user_id": userId})
}

func (db Database) MemberCreate(member model_database.Member) bool {
	return Insert(db, TableMembers, member) != nil
}

func (db Database) MemberDelete(userId, groupId uint64) bool {
	return Delete(db, TableMembers, exp.Ex{"group_id": groupId, "user_id": userId})
}

func (db Database) MemberList(groupId uint64) []model_database.User {
	memberList := new([]model_database.User)
	query := `SELECT ` + TableUsers + `.*
	FROM ` + TableMembers + `
	LEFT JOIN ` + TableUsers + ` ON ` + TableUsers + `.id = ` + TableMembers + `.user_id
	WHERE ` + TableMembers + `.group_id = ?`
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
	memberList := new([]model_database.User)
	query := `SELECT ` + TableUsers + `.* FROM ` + TableMembers + `
	LEFT JOIN ` + TableUsers + ` ON ` + TableUsers + `.id = ` + TableMembers + `.user_id
	WHERE ` + TableMembers + `.group_id = ?
	ORDER BY ` + TableMembers + `.user_id ASC LIMIT ?, ?`
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
