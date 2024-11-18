package internal

import "restapp/internal/model"

// Create new DB record.
func (db Database) GroupCreate(group model.Group) error {
	query :=
		`INSERT INTO app_groups (
			nickname,
			groupname,
			groupmode,
			password,
			avatar,
			created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
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
    	SET nickname = ?,
		groupname = ?,
		groupmode = ?,
		password = ?,
		avatar = ?,
		created_at = ?

        WHERE id = ?`
	_, err := db.Sql.Exec(query,
		group.Nick,
		group.Name,
		group.Password,
		group.Avatar,
		group.CreatedAt,
		group.Id,
	)
	return err
}

// Delete the existing DB record.
func (db Database) GroupDelete(id uint) error {
	query := `DELETE FROM app_groups WHERE id = ?`
	_, err := db.Sql.Exec(query, id)
	// TODO: delete roles, members, messages, add description ^^^
	return err
}

// Get the group by her identificator.
func (db Database) GroupById(id uint) (*model.Group, error) {
	group := new(model.Group)
	query := `SELECT * FROM app_groups WHERE id = ?`
	err := db.Sql.Get(group, query, id)
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
