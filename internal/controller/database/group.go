package database

import (
	"github.com/Mopsgamer/vibely/internal/controller/model_database"
)

// Create new DB record.
func (db Database) GroupCreate(group model_database.Group) *uint64 {
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
		logSqlError(err)
		return nil
	}

	newId := &db.Context().LastInsertId
	return newId
}

// Change the existing DB record.
func (db Database) GroupUpdate(group model_database.Group) bool {
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
		logSqlError(err)
		return false
	}
	return true
}

// Delete the existing DB record and memberships.
func (db Database) GroupDelete(groupId uint64) bool {
	query := `DELETE FROM app_groups WHERE id = ?`
	_, err := db.Sql.Exec(query, groupId)
	if err != nil {
		logSqlError(err)
		return false
	}

	query = `DELETE FROM app_group_members WHERE group_id = ?`
	_, err = db.Sql.Exec(query, groupId)
	if err != nil {
		logSqlError(err)
		return false
	}

	// TODO: delete roles(!) and messages(?) when deleting group

	return true
}

// Get the group by her identificator.
func (db Database) GroupById(groupId uint64) *model_database.Group {
	group := new(model_database.Group)
	query := `SELECT * FROM app_groups WHERE id = ?`
	err := db.Sql.Get(group, query, groupId)

	if err != nil {
		logSqlError(err)
		return nil
	}
	return group
}

// Get the group by her group name.
func (db Database) GroupByName(groupName string) *model_database.Group {
	group := new(model_database.Group)
	query := `SELECT * FROM app_groups WHERE groupname = ?`
	err := db.Sql.Get(group, query, groupName)

	if err != nil {
		logSqlError(err)
		return nil
	}
	return group
}
