package model

import "time"

type Group struct {
	Id   string `db:"id"`
	Nick string `db:"nickname"`
	Name string `db:"groupname"`
	Mode string `db:"groupmode"`
	// Hashed password string
	Password  *string   `db:"password"`
	Avatar    string    `db:"avatar"`
	CreatedAt time.Time `db:"created_at"`
}
