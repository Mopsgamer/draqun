package internal

import (
	"restapp/internal/model"

	"github.com/jmoiron/sqlx"
)

// The SQL DB wrapper.
type Database struct{ Sql *sqlx.DB }

// Create new DB record.
func (db Database) UserCreate(user model.User) error {
	query := `INSERT INTO users (name, tag, email, phone, password, avatar, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query, user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt)
	return err
}

// Change the existing DB record.
func (db Database) UserUpdate(user model.User) error {
	query := `UPDATE users 
              SET name = ?, tag = ?, email = ?, phone = ?, password = ?, avatar = ?, created_at = ? 
              WHERE id = ?`
	_, err := db.Sql.Exec(query, user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt, user.ID)
	return err
}

// Delete the existing DB record.
func (db Database) DeleteUser(user model.User) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := db.Sql.Exec(query, user.ID)
	return err
}

// Get the user by his email.
func (db Database) UserByEmail(email string) (*model.User, error) {
	user := new(model.User)
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
	FROM users WHERE email = ?`
	err := db.Sql.Get(user, query, email)
	if err != nil {
		user = nil
	}
	return user, err
}

// Get the user by his identificator.
func (db Database) UserByID(userID int) (*model.User, error) {
	user := new(model.User)
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
	FROM users WHERE id = ?`
	err := db.Sql.Get(user, query, userID)
	if err != nil {
		user = nil
	}
	return user, err
}
