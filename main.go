package main

import (
	internal "github.com/Mopsgamer/draqun/server"
	"github.com/Mopsgamer/draqun/server/environment"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3/log"
)

func main() {
	environment.Load()
	if app, err := internal.NewApp(); err == nil {
		log.Fatal(app.Listen(":" + environment.Port))
	}
}
