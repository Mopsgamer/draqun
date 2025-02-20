package model_database

import (
	"time"
)

// User token expiration: 24 hours.
var UserTokenExpiration = 24 * time.Hour

// The user as a database entry
type User struct {
	Id uint64 `db:"id"`
	// Nick is customizable name. Can contain emojis and special characters.
	Nick string `db:"nickname"`
	// Name is a simple identificator, which can be used to create friend links.
	Name  string  `db:"username"`
	Email string  `db:"email"`
	Phone *string `db:"phone"`
	// Hashed password string.
	Password  string    `db:"password"`
	Avatar    string    `db:"avatar"`
	CreatedAt time.Time `db:"created_at"`
	LastSeen  time.Time `db:"last_seen"`
}

// func (user *User) IsOnline() bool {
// 	if user == nil {
// 		return false
// 	}

// 	cons := *controller_ws.UserSessionMap.mp
// 	arr, ok := cons[user.Id]
// 	return ok && len(arr) > 0
// }
