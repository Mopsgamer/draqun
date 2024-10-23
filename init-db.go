package main

import (
	"fmt"
	"log"
	"os"
	"restapp/restapp"

	"github.com/jmoiron/sqlx"
)

func InitDB() (*restapp.Database, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	connection, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		return nil, err
	}

	if err := connection.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
		return nil, err
	}

	log.Println("Database connected successfully")
	return &restapp.Database{Sql: connection}, nil
}
