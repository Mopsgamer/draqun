package database

import (
	"restapp/internal/logic/model_database"

	"github.com/gofiber/fiber/v3/log"
)

func (db Database) RoleCreate(right model_database.Role) bool {
	query :=
		`INSERT INTO app_group_roles (
			group_id,
			user_id,
			right_id
		)
    	VALUES (?, ?, ?)`
	_, err := db.Sql.Exec(query,
		right.GroupId,
		right.UserId,
		right.RightId,
	)

	if err != nil {
		log.Error(err)
		return false
	}

	return true
}
