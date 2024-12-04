package database

import (
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

// Create new DB record.
func (db Database) UserCreate(user model_database.User) *uint64 {
	query :=
		`INSERT INTO app_users (
			nickname,
			username,
			email,
			phone,
			password,
			avatar,
			created_at,
			last_seen
		)
    	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		user.Nick,
		user.Name,
		user.Email,
		user.Phone,
		user.Password,
		user.Avatar,
		user.CreatedAt,
		user.LastSeen,
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := &db.Context().LastInsertId
	return newId
}

// Change the existing DB record.
func (db Database) UserUpdate(user model_database.User) bool {
	query :=
		`UPDATE app_users
    	SET
		nickname = ?,
		username = ?,
		email = ?,
		phone = ?,
		password = ?,
		avatar = ?,
		created_at = ?,
		last_seen = ?

        WHERE id = ?`
	_, err := db.Sql.Exec(query,
		user.Nick,
		user.Name,
		user.Email,
		user.Phone,
		user.Password,
		user.Avatar,
		user.CreatedAt,
		user.LastSeen,
		user.Id,
	)

	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

// Delete the existing DB record.
func (db Database) UserDelete(userId uint64) bool {
	query := `DELETE FROM app_users WHERE id = ?`
	_, err := db.Sql.Exec(query, userId)

	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

// Get the user by his email.
func (db Database) UserByEmail(email string) *model_database.User {
	user := new(model_database.User)
	query := `SELECT * FROM app_users WHERE email = ?`
	err := db.Sql.Get(user, query, email)

	if err != nil {
		log.Error(err)
		return nil
	}
	return user
}

// Get the user by his identificator.
func (db Database) UserById(userId uint64) *model_database.User {
	user := new(model_database.User)
	query := `SELECT * FROM app_users WHERE id = ?`
	err := db.Sql.Get(user, query, userId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return user
}

// Get the user by his username.
func (db Database) UserByUsername(username string) *model_database.User {
	user := new(model_database.User)
	query := `SELECT * FROM app_users WHERE username = ?`
	err := db.Sql.Get(user, query, username)

	if err != nil {
		log.Error(err)
		return nil
	}
	return user
}

func (db Database) UserRoleList(groupId, userId uint64) []model_database.Role {
	roleList := &[]model_database.Role{}
	query := `SELECT * FROM app_group_roles WHERE group_id = ? AND user_id = ?`
	err := db.Sql.Select(roleList, query, groupId, userId)

	if err != nil {
		log.Error(err)
		return *roleList
	}
	return *roleList
}

func (db Database) RightById(rightId uint64) *model_database.Right {
	right := new(model_database.Right)
	query := `SELECT * FROM app_group_role_rights WHERE id = ?`
	err := db.Sql.Get(right, query, rightId)

	if err != nil {
		log.Error(err)
		return right
	}
	return right
}

func (db Database) UserRightsList(groupId, userId uint64) []model_database.Right {
	rightList := new([]model_database.Right)
	query := `SELECT app_group_role_rights.*
	FROM app_group_role_rights
	LEFT JOIN app_group_roles ON app_group_roles.right_id = app_group_role_rights.id
	WHERE app_group_role_rights.group_id = ? AND app_group_role_rights.user_id = ?`
	err := db.Sql.Select(rightList, query, groupId, userId)

	if err != nil {
		log.Error(err)
		return *rightList
	}
	return *rightList
}

func (db Database) UserRights(groupId, userId uint64) model_database.Right {
	roleList := db.UserRightsList(groupId, userId)
	rights := model_database.Right{}
	if len(roleList) == 0 {
		return rights
	}

	rights.Merge(roleList...)
	return rights
}

func (db Database) UserOwnGroupList(userId uint64) []model_database.Group {
	groupList := &[]model_database.Group{}
	query := `SELECT
		app_groups.*
	FROM app_groups
	LEFT JOIN app_group_members ON app_groups.id = app_group_members.group_id
	WHERE (user_id = ? AND is_owner = 1)`
	err := db.Sql.Select(groupList, query, userId)

	if err != nil {
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
	err := db.Sql.Select(groupList, query, userId)

	if err != nil {
		log.Error(err)
		return *groupList
	}
	return *groupList
}

func (db Database) UserJoinGroup(newMember model_database.Member) bool {
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
		newMember.GroupId,
		newMember.UserId,
		newMember.IsOwner,
		newMember.IsBanned,
		newMember.Nick,
	)

	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (db Database) UserLeaveGroup(userId, groupId uint64) bool {
	query := `DELETE FROM app_group_members WHERE group_id = ? AND user_id = ?`
	_, err := db.Sql.Exec(query, groupId, userId)

	if err != nil {
		log.Error(err)
		return false
	}
	return true
}

func (db Database) MessageById(messageId uint64) *model_database.Message {
	message := new(model_database.Message)
	query := `SELECT * FROM app_group_role_rights WHERE id = ?`
	err := db.Sql.Get(message, query, messageId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return message
}

func (db Database) MessageCreate(message model_database.Message) *uint64 {
	query :=
		`INSERT INTO app_group_messages (
			group_id,
			author_id,
			content,
			created_at
		)
    	VALUES (?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		message.GroupId,
		message.AuthorId,
		message.Content,
		message.CreatedAt,
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := &db.Context().LastInsertId
	return newId
}
