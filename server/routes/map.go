package routes

import (
	"errors"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/Mopsgamer/draqun/server/model"
	"github.com/Mopsgamer/draqun/server/perms"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func MapPage(ctx fiber.Ctx, bind fiber.Map) fiber.Map {
	// empty values
	fiber.Locals(ctx, perms.LocalAuth, model.User{})
	// actually fill
	_, err := perms.UserByAuthFromCtx(ctx)
	if err != nil && errors.Is(err, htmx.ErrToken) {
		log.Error(err)
	}

	// other values
	bindx := fiber.Map{
		"Ctx": ctx,

		"AppName":      environment.AppName,
		"GitHubRepo":   environment.GitHubRepo,
		"GitHubCommit": environment.GitHubCommit,
		"GitHubBranch": environment.GitHubBranch,
		"DenoJson":     environment.DenoJson,
		"GitJson":      environment.GitJson,
		"GoMod":        environment.GoMod,
	}

	path := ctx.Path()

	var stylesheets []string
	var scripts []string

	if strings.HasPrefix(path, "/chat") {
		stylesheets = []string{"/static/css/main.css"}
		scripts = []string{"/static/js/app.js"}
	} else if strings.HasPrefix(path, "/docs") {
		stylesheets = []string{"/static/css/docs.css"}
		scripts = []string{"/static/js/docs.js"}
	} else if path == "/" {
		stylesheets = []string{"/static/css/homepage.css"}
		scripts = []string{"/static/js/homepage.js"}
	} else {
		stylesheets = []string{"/static/css/main.css"}
		scripts = []string{"/static/js/main.js"}
	}

	if environment.BuildEnvironment == environment.BuildModeDevelopment {
		version := strconv.FormatInt(time.Now().UnixNano(), 10)
		for i, stylesheet := range stylesheets {
			stylesheets[i] = stylesheet + "?v=" + version
		}
		for i, script := range scripts {
			scripts[i] = script + "?v=" + version
		}
	}

	bindx["Stylesheets"] = stylesheets
	bindx["Scripts"] = scripts

	maps.Copy(bindx, bind)
	return bindx
}
