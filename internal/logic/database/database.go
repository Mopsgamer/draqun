package database

import (
	"fmt"
	"restapp/internal/environment"

	"github.com/gofiber/fiber/v3/log"
	"github.com/jmoiron/sqlx"
)

func New(sqlDB *sqlx.DB) Database {
	return Database{sqlDB}
}

// The SQL DB wrapper.
type Database struct {
	Sql *sqlx.DB
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

	log.Info("Database connected successfully. Hope she is set up manually or by 'deno task init'.")
	return &Database{Sql: connection}, nil
}

type DatabaseContext struct {
	// Not each id is uint64 (BIGINT) so you should convert it to uint32 (MEDIUMINT).
	LastInsertId uint64 `db:"new_id"`
}

func (db Database) Context() *DatabaseContext {
	context := new(DatabaseContext)
	err := db.Sql.Get(context, `SELECT LAST_INSERT_ID() AS new_id`)
	if err != nil {
		log.Error(err)
		return nil
	}

	return context
}
