package internal

import (
	"restapp/internal/model"

	"github.com/gofiber/fiber/v3/log"
)

// Create new DB record.
func (db Database) UserCreate(user model.User) *uint64 {
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
func (db Database) UserUpdate(user model.User) bool {
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
func (db Database) UserByEmail(email string) *model.User {
	user := new(model.User)
	query := `SELECT * FROM app_users WHERE email = ?`
	err := db.Sql.Get(user, query, email)

	if err != nil {
		log.Error(err)
		return nil
	}
	return user
}

// Get the user by his identificator.
func (db Database) UserById(userId uint64) *model.User {
	user := new(model.User)
	query := `SELECT * FROM app_users WHERE id = ?`
	err := db.Sql.Get(user, query, userId)

	if err != nil {
		log.Error(err)
		return nil
	}
	return user
}

// Get the user by his username.
func (db Database) UserByUsername(username string) *model.User {
	user := new(model.User)
	query := `SELECT * FROM app_users WHERE username = ?`
	err := db.Sql.Get(user, query, username)

	if err != nil {
		log.Error(err)
		return nil
	}
	return user
}

func (db Database) UserRoleList(userId uint64) []model.Role {
	roleList := &[]model.Role{}
	query := `SELECT * FROM app_group_roles WHERE user_id = ?`
	err := db.Sql.Select(roleList, query, userId)

	if err != nil {
		log.Error(err)
		return *roleList
	}
	return *roleList
}

func (db Database) RoleRights(rightId uint64) *model.Rights {
	roleList := new(model.Rights)
	query := `SELECT * FROM app_group_role_rights WHERE id = ?`
	err := db.Sql.Get(roleList, query, rightId)

	if err != nil {
		log.Error(err)
		return roleList
	}
	return roleList
}

func (db Database) UserRights(userId uint64) *model.Rights {
	rights := new(model.Rights)
	// FIXME: user rights sql query
	// since the user can have multiple roles,
	// we should calculate it as a single right object.
	query := `SELECT * FROM app_group_role_rights WHERE id = ?`
	err := db.Sql.Get(rights, query, userId)

	if err != nil {
		log.Error(err)
		return rights
	}
	return rights
}

func (db Database) UserOwnGroupList(userId uint64) []model.Group {
	groupList := &[]model.Group{}
	query := `SELECT
		id, creator_id, nickname, groupname, groupmode, description, password, avatar, created_at
	FROM app_groups
	LEFT JOIN app_group_members ON app_groups.id = app_group_members.group_id
	WHERE (user_id = ? AND is_owner = 1)
	GROUP BY app_group_members.group_id`
	err := db.Sql.Select(groupList, query, userId)

	if err != nil {
		log.Error(err)
		return *groupList
	}
	return *groupList
}

func (db Database) UserGroupList(userId uint64) []model.Group {
	groupList := &[]model.Group{}
	query := `SELECT
		id, creator_id, nickname, groupname, groupmode, description, password, avatar, created_at
	FROM app_groups
	LEFT JOIN app_group_members ON app_groups.id = app_group_members.group_id
	WHERE user_id = ?
	GROUP BY app_group_members.group_id`
	err := db.Sql.Select(groupList, query, userId)

	if err != nil {
		log.Error(err)
		return *groupList
	}
	return *groupList
}

func (db Database) UserJoinGroup(newMember model.Member) bool {
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

func (db Database) MessageCreate(message model.Message) *uint64 {
	query :=
		`INSERT INTO app_group_messages (
			group_id,
			author_id,
			content,
			created_at
		)
    	VALUES (?, ?, ?, ?, ?)`
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
