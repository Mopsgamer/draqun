package internal

import (
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/gofiber/fiber/v3/log"
)

func Serve(embedFS fs.FS, clientEmbedded bool) {
	clientEmbeddedStatus := "client not embedded"
	if clientEmbedded {
		clientEmbeddedStatus = "client embedded"
	}

	environment.LoadMeta(embedFS)
	log.Infof("Server version: v%s (%s %s), %s, %s", environment.DenoJson.Version, environment.GitJson.Branch, environment.GitJson.Hash, clientEmbeddedStatus, environment.BuildModeName)
	environment.LoadEnv(embedFS)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Info("Served!")
		os.Exit(0)
	}()

	if app, err := NewApp(embedFS, clientEmbedded); err == nil {
		err = app.Listen(":" + environment.Port) // normal

		if err == nil {
			return
		}

		if environment.BuildModeValue == environment.BuildModeProduction {
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
