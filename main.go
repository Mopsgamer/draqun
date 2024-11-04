package main

import (
	"log"
	"restapp/internal"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	if internal.InitProject() {
		return
	}
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		return
	}
	if app, err := internal.InitServer(); err == nil {
		log.Fatal(app.Listen(":3000"))
	}
}
