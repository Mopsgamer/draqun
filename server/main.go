package internal

import (
	"fmt"
	"io/fs"

	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/gofiber/fiber/v3"
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
	model.LoadDB()
	if err != nil {
		fmt.Println(meta)
		log.Fatal(err)
	}

	app, err := NewApp(embedFS, clientEmbedded)
	if err != nil {
		log.Fatal(err)
	}

	app.Hooks().OnListen(func(data fiber.ListenData) error {
		fmt.Println(meta)
		return nil
	})

	err = app.Listen(":" + environment.Port) // normal
	if err == nil {
		return
	}

	if environment.BuildModeValue == environment.BuildModeProduction {
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

func metaString(clientEmbedded bool) string {
	cliembColor := color.RGB(0, 180, 100)
	clientEmbeddedStatus := "client not embedded"
	if clientEmbedded {
		cliembColor = color.New(color.FgHiRed)
		clientEmbeddedStatus = "client embedded"
	}

	hashColor := color.New(color.Faint)
	branchColor := color.RGB(100, 0, 180)
	version := branchColor.Sprint(link(environment.GitHubBranch, environment.GitJson.Branch)) + " " +
		"v" + environment.DenoJson.Version + " " +
		hashColor.Sprint(link(environment.GitHubCommit, environment.GitJson.Hash))

	prefix := "│  "
	title := environment.AppName + " ─ "
	return "╭── " + color.New(color.Italic).Sprint(title) + "\n" +
		prefix + "\n" +
		prefix + version + "\n" +
		prefix + color.HiRedString(environment.BuildModeName) + "\n" +
		prefix + cliembColor.Sprint(clientEmbeddedStatus) + "\n" +
		prefix
}

func link(link, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", link, text)
}
