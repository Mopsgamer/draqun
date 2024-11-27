package internal

import (
	"restapp/internal/model"

	"github.com/gofiber/fiber/v3/log"
)

// Create new DB record.
func (db Database) GroupCreate(group model.Group) *uint64 {
	query :=
		`INSERT INTO app_groups (
			creator_id,
			nickname,
			groupname,
			groupmode,
			description,
			password,
			avatar,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		group.CreatorId,
		group.Nick,
		group.Name,
		group.Mode,
		group.Description,
		group.Password,
		group.Avatar,
		group.CreatedAt,
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := &db.Context().LastInsertId
	return newId
}

// Change the existing DB record.
func (db Database) GroupUpdate(group model.Group) bool {
	query :=
		`UPDATE app_groups
    	SET
		creator_id = ?,
		nickname = ?,
		groupname = ?,
		groupmode = ?,
		description = ?,
		password = ?,
		avatar = ?,
		created_at = ?

        WHERE id = ?`
	_, err := db.Sql.Exec(query,
		group.CreatorId,
		group.Nick,
		group.Name,
		group.Mode,
		group.Description,
		group.Password,
		group.Avatar,
		group.CreatedAt,
		group.Id,
	)

	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

// Delete the existing DB record and memberships.
func (db Database) GroupDelete(groupId uint64) bool {
	query := `DELETE FROM app_groups WHERE id = ?`
	_, err := db.Sql.Exec(query, groupId)
	if err != nil {
		log.Error(err)
		return false
	}

	query = `DELETE FROM app_group_members WHERE group_id = ?`
	_, err = db.Sql.Exec(query, groupId)
	if err != nil {
		log.Error(err)
		return false
	}

	// TODO: delete roles(!) and messages(?) when deleting group

	return true
}

// Get the group by her identificator.
func (db Database) GroupById(groupId uint64) *model.Group {
	group := new(model.Group)
	query := `SELECT * FROM app_groups WHERE id = ?`
	err := db.Sql.Get(group, query, groupId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return group
}

// Get the group by her groupname.
func (db Database) GroupByGroupname(groupname string) *model.Group {
	group := new(model.Group)
	query := `SELECT * FROM app_groups WHERE groupname = ?`
	err := db.Sql.Get(group, query, groupname)

	if err != nil {
		log.Error(err)
		return nil
	}
	return group
}

func (db Database) GroupMemberList(groupId uint64) []model.User {
	memberList := &[]model.User{}
	query := `SELECT
		app_users.*
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

func (db Database) GroupMessageList(groupId uint64) []model.Message {
	messageList := &[]model.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ?`
	err := db.Sql.Select(messageList, query, groupId)

	if err != nil {
		log.Error(err)
		return *messageList
	}
	return *messageList
}

func (db Database) GroupMemberById(groupId, userId uint64) *model.Member {
	member := new(model.Member)
	query := `SELECT * FROM app_group_members WHERE user_id = ?`
	err := db.Sql.Get(member, query, userId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return member
}

func (db Database) GroupMemberCreate(member model.Member) *uint64 {
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

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := &db.Context().LastInsertId
	return newId
}
