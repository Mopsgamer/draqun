package internal

import (
	"errors"
	"log"

	"github.com/jmoiron/sqlx"
)

type Database struct{ Sql *sqlx.DB }

func (db Database) UserSave(user User) error {
	query := `INSERT INTO users (name, tag, email, phone, password, avatar, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query, user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt)
	if err != nil {
		log.Println(err)
		log.Println(user)
		return err
	}
	return nil
}

func (d Database) UserByEmail(email string) (*User, error) {
	var user = new(User)
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
              FROM users WHERE email = ?`
	err := d.Sql.Get(&user, query, email)
	if err != nil {
		log.Println(err)
		log.Println(user)
		return nil, errors.New("user not found")
	}
	return user, nil
}
