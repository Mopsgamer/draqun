package main

import (
	"restapp/internal"
	"restapp/internal/environment"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3/log"
)

func main() {
	environment.Load()
	if app, err := internal.NewApp(); err == nil {
		log.Fatal(app.Listen(":3000"))
	}
}
