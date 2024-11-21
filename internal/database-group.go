package internal

import "restapp/internal/model"

// Create new DB record.
func (db Database) GroupCreate(group model.Group) error {
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
	return err
}

// Change the existing DB record.
func (db Database) GroupUpdate(group model.Group) error {
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
	return err
}

// Delete the existing DB record and memberships.
func (db Database) GroupDelete(groupId uint) error {
	query := `DELETE FROM app_groups WHERE id = ?`
	_, err := db.Sql.Exec(query, groupId)
	if err != nil {
		return err
	}

	query = `DELETE FROM app_group_members WHERE group_id = ?`
	_, err = db.Sql.Exec(query, groupId)
	if err != nil {
		return err
	}

	// TODO: delete roles(!) and messages(?)

	return err
}

// Get the group by her identificator.
func (db Database) GroupById(groupId uint) (*model.Group, error) {
	group := new(model.Group)
	query := `SELECT * FROM app_groups WHERE id = ?`
	err := db.Sql.Get(group, query, groupId)
	if err != nil {
		group = nil
	}
	return group, err
}

// Get the group by her groupname.
func (db Database) GroupByGroupname(groupname string) (*model.Group, error) {
	group := new(model.Group)
	query := `SELECT * FROM app_groups WHERE groupname = ?`
	err := db.Sql.Get(group, query, groupname)
	if err != nil {
		group = nil
	}
	return group, err
}

func (db Database) GroupMembers(groupId uint) (*[]model.Member, error) {
	memberList := new([]model.Member)
	query := `SELECT * FROM app_group_members WHERE group_id = ?`
	err := db.Sql.Get(memberList, query, groupId)
	if err != nil {
		memberList = nil
	}
	return memberList, err
}
