package internal

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
)

// The SQL DB wrapper.
type Database struct{ Sql *sqlx.DB }

// Save the user as a DB record.
func (db Database) UserSave(user User) error {
	query := `INSERT INTO users (name, tag, email, phone, password, avatar, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query, user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt)
	if err != nil {
		log.Error(err)
		log.Info(user)
		return err
	}
	return nil
}

// Get the user by his email.
func (db Database) UserByEmail(email string) (*User, error) {
	email = strings.TrimSpace(email)
	var user = new(User)
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
              FROM users WHERE email = ?`
	err := db.Sql.Get(user, query, email)
	if err != nil {
		log.Error(err)
		log.Info(user)
		return nil, errors.New("User not found")
	}
	return user, nil
}

// Get the user by his identificator.
func (db Database) UserByID(userID int) (*User, error) {
	var user = new(User)
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
              FROM users WHERE id = ?`
	err := db.Sql.Get(user, query, userID)
	if err != nil {
		log.Error(err)
		return nil, errors.New("User not found")
	}
	return user, nil
}
