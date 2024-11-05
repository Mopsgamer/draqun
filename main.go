package main

import (
	"restapp/internal"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
	if app, err := internal.InitServer(); err == nil {
		log.Fatal(app.Listen(":3000"))
	}
}
