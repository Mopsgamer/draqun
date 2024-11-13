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

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
		name VARCHAR(255) DEFAULT NULL,
		tag VARCHAR(255) DEFAULT NULL,
		email VARCHAR(255) NOT NULL,
		phone VARCHAR(255) DEFAULT NULL,
		password VARCHAR(255) NOT NULL,
		avatar VARCHAR(255) DEFAULT NULL,
		created_at DATETIME DEFAULT NULL COMMENT 'Account create time',
		last_seen DATETIME DEFAULT NULL COMMENT 'Last seen time',
		registered TINYINT(1) DEFAULT '0',
		PRIMARY KEY (id)
	) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Users data'
`

	if _, err := connection.Exec(createTableQuery); err != nil {
		return nil, err
	}

	log.Info("Users table ensured to exist")
	return &Database{Sql: connection}, nil
}
