package model_database

import (
	"time"
)

// The user as a database entry
type User struct {
	Id        uint64    `db:"id"`
	Nick      string    `db:"nickname"` // Nick is customizable name. Can contain emojis and special characters.
	Name      string    `db:"username"` // Name is a simple identificator, which can be used to create friend links.
	Email     string    `db:"email"`
	Phone     *string   `db:"phone"`
	Password  string    `db:"password"` // Hashed password string.
	Avatar    string    `db:"avatar"`
	CreatedAt time.Time `db:"created_at"`
	LastSeen  time.Time `db:"last_seen"`
}
