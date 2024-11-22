package internal

import (
	"restapp/internal/model"

	"github.com/gofiber/fiber/v3/log"
)

// Create new DB record.
func (db Database) GroupCreate(group model.Group) bool {
	query :=
		`INSERT INTO app_groups (
			creator_id,
			nickname,
			groupname,
			groupmode,
			password,
			avatar,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		group.CreatorId,
		group.Nick,
		group.Name,
		group.Mode,
		group.Password,
		group.Avatar,
		group.CreatedAt,
	)

	if err != nil {
		log.Error(err)
		return false
	}
	return true
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
		password = ?,
		avatar = ?,
		created_at = ?

        WHERE id = ?`
	_, err := db.Sql.Exec(query,
		group.CreatorId,
		group.Nick,
		group.Name,
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
func (db Database) GroupDelete(groupId uint) bool {
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

	// TODO: delete roles(!) and messages(?)

	return true
}

// Get the group by her identificator.
func (db Database) GroupById(groupId uint) *model.Group {
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

func (db Database) GroupMemberList(groupId uint) []model.Member {
	memberList := []model.Member{}
	query := `SELECT * FROM app_group_members WHERE group_id = ?`
	err := db.Sql.Get(memberList, query, groupId)

	if err != nil {
		log.Error(err)
		return memberList
	}
	return memberList
}

func (db Database) GroupMessageList(groupId uint) []model.Message {
	messageList := []model.Message{}
	query := `SELECT * FROM app_group_messages WHERE group_id = ?`
	err := db.Sql.Get(messageList, query, groupId)

	if err != nil {
		log.Error(err)
		return messageList
	}
	return messageList
}

func (db Database) GroupMember(groupId, userId uint) *model.Member {
	member := new(model.Member)
	query := `SELECT * FROM app_group_members WHERE user_id = ?`
	err := db.Sql.Get(member, query, userId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return member
}
