package internal

import (
	"errors"
	"restapp/internal/model"
)

// Create new DB record.
func (db Database) UserCreate(user model.User) error {
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
	return err
}

// Change the existing DB record.
func (db Database) UserUpdate(user model.User) error {
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
	return err
}

// Delete the existing DB record.
func (db Database) UserDelete(id uint) error {
	query := `DELETE FROM app_users WHERE id = ?`
	_, err := db.Sql.Exec(query, id)
	return err
}

// Get the user by his email.
func (db Database) UserByEmail(email string) (*model.User, error) {
	user := new(model.User)
	query := `SELECT * FROM app_users WHERE email = ?`
	err := db.Sql.Get(user, query, email)
	if err != nil {
		user = nil
	}
	return user, err
}

// Get the user by his identificator.
func (db Database) UserById(id uint) (*model.User, error) {
	user := new(model.User)
	query := `SELECT * FROM app_users WHERE id = ?`
	err := db.Sql.Get(user, query, id)
	if err != nil {
		user = nil
	}
	return user, err
}

// Get the user by his username.
func (db Database) UserByUsername(username string) (*model.User, error) {
	user := new(model.User)
	query := `SELECT * FROM app_users WHERE username = ?`
	err := db.Sql.Get(user, query, username)
	if err != nil {
		user = nil
	}
	return user, err
}

func (db Database) UserJoinGroup(userId uint, groupId uint) error {
	// TODO: sql - join group
	// check group mode
	return errors.New("not implemented")
}

func (db Database) UserLeaveGroup(userId uint, groupId uint) error {
	// TODO: sql - leave group
	return errors.New("not implemented")
}

func (db Database) UserSendMessage(userId uint, groupId uint, message model.Message) error {
	// TODO: sql - send message
	return errors.New("not implemented")
}
