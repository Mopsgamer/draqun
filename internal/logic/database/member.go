package database

import (
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

func (db Database) MemberList(groupId uint64) []model_database.User {
	memberList := &[]model_database.User{}
	query := `SELECT app_users.*
	FROM app_group_members
	LEFT JOIN app_users ON app_users.id = app_group_members.user_id
	WHERE app_group_members.group_id = ?`
	err := db.Sql.Select(memberList, query, groupId)

	if err != nil {
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
	err := db.Sql.Select(memberList, query, groupId, from, to)
	if err != nil {
		log.Error(err)
		return *memberList
	}
	return *memberList
}

func (db Database) MemberById(groupId, userId uint64) *model_database.Member {
	member := new(model_database.Member)
	query := `SELECT * FROM app_group_members WHERE user_id = ?`
	err := db.Sql.Get(member, query, userId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return member
}

func (db Database) MemberCreate(member model_database.Member) error {
	query :=
		`INSERT INTO app_group_members (
			group_id,
			user_id,
			is_owner,
			is_banned,
			membernick
		)
		VALUES (?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		member.GroupId,
		member.UserId,
		member.IsOwner,
		member.IsBanned,
		member.Nick,
	)

	return err
}
