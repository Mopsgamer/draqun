package database

import (
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

func (db Database) RightCreate(right model_database.Right) *uint32 {
	query :=
		`INSERT INTO app_group_role_rights (
			group_id,
			perm_chat_read,
			perm_chat_write,
			perm_chat_delete,
			perm_kick,
			perm_ban,
			perm_change_group,
			perm_change_member
		)
    	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		right.GroupId,
		right.ChatRead,
		right.ChatWrite,
		right.ChatDelete,
		right.Kick,
		right.Ban,
		right.ChangeGroup,
		right.ChangeMember,
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := uint32(db.Context().LastInsertId)
	return &newId
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
