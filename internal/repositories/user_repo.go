package repository

import (
	"errors"

	"restapp/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	CreateUser(user models.User) error
	GetUserByEmail(email string) (*models.User, error)
}

type DBUserRepository struct {
	db *sqlx.DB
}

func NewDBUserRepository(db *sqlx.DB) *DBUserRepository {
	return &DBUserRepository{db: db}
}

func (repo *DBUserRepository) CreateUser(user models.User) error {
	query := `INSERT INTO users (name, tag, email, phone, password, avatar, created_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := repo.db.Exec(query, user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt)
	/*log.Printf("%s, %s, %s, %s, %s, %s, %s\nError: %s",
	user.Name, user.Tag, user.Email, user.Phone, user.Password, user.Avatar, user.CreatedAt, err)*/
	if err != nil {
		return err
	}
	return nil
}

func (repo *DBUserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := `SELECT id, name, tag, email, phone, password, avatar, created_at 
              FROM users WHERE email = ?`
	err := repo.db.Get(&user, query, email)
	//log.Printf("Error: %s", err)
	//log.Println(user)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
