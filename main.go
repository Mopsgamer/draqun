package main

import (
	"embed"

	server "github.com/Mopsgamer/draqun/server"
	"github.com/Mopsgamer/draqun/server/environment"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v3/log"
)

//go:embed client/static/** client/templates/**
var embedFS embed.FS

func main() {
	environment.Load()
	if app, err := server.NewApp(nil); err == nil {
		err = app.Listen(":" + environment.Port) // normal

		if err == nil {
			return
		}

		if environment.Environment == environment.BuildModeProduction {
			log.Fatal(err)
			return
		}

		switch environment.Port {
		case "3000":
			environment.Port = "8080"
		case "8080":
			environment.Port = "3000"
		default:
			environment.Port = "0"
		}
		log.Fatal(app.Listen(":" + environment.Port)) // fallback
	}
}
