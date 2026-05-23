package internal

import (
	"fmt"
	"io/fs"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3/log"
)

func Serve(embedFS fs.FS, clientEmbedded bool) {
	environment.LoadMeta(embedFS)
	meta := metaString(clientEmbedded)

	err := environment.LoadEnv(embedFS)
	if err != nil {
		fmt.Println(meta)
		log.Fatal(err)
	}
	err = model.LoadDB()
	if err != nil {
		fmt.Println(meta)
		log.Fatal(err)
	}

	app, err := NewApp(embedFS, clientEmbedded)
	if err != nil {
		log.Fatal(err)
	}

	err = app.Listen(":" + environment.Port) // normal
	if err == nil {
		return
	}

	if environment.BuildEnvironment == environment.BuildModeProduction {
		log.Fatal(err)
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
