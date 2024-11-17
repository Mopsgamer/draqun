package internal

import (
	"fmt"
	"restapp/internal/environment"

	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
)

// The SQL DB wrapper.
type Database struct{ Sql *sqlx.DB }

// Initialize the DB wrapper.
func InitDB() (*Database, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		environment.DBUser,
		environment.DBPassword,
		environment.DBHost,
		environment.DBPort,
		environment.DBName,
	)

	connection, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	if err := connection.Ping(); err != nil {
		return nil, err
	}

	log.Info("Database connected successfully")
	log.Info("Hope she is set up manually or by 'deno task init'")
	return &Database{Sql: connection}, nil
}
