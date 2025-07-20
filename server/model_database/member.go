package model_database

import (
	"time"
)

type Member struct {
	GroupId           uint64    `db:"group_id"`
	UserId            uint64    `db:"user_id"`
	Moniker           string    `db:"moniker"`
	FirstTimeJoinedAt time.Time `db:"first_time_joined_at"`
}
