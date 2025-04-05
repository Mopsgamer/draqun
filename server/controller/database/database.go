package database

import (
	"fmt"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/doug-martin/goqu/v9"

	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
)

// The SQL DB wrapper.
type Database struct {
	Sqlx *sqlx.DB
	Goqu *goqu.Database
}

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

	goquConnection := goqu.New("mysql", connection)
	log.Info("Database connected successfully. Hope she is set up manually or by 'deno task init'.")
	return &Database{Sqlx: connection, Goqu: goquConnection}, nil
}
