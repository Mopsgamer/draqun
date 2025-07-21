package controller

import (
	"bytes"
	"html/template"

	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

func MapPage(ctx fiber.Ctx, bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{
		"AppName":     environment.AppName,
		"GitHubRepo":  environment.GitHubRepo,
		"DenoJson":    environment.DenoJson,
		"GoMod":       environment.GoMod,
		"GitHash":     environment.GitHash,
		"GitHashLong": environment.GitHashLong,

		"User":   fiber.Locals[database.User](ctx, LocalAuth),
		"Group":  fiber.Locals[database.Group](ctx, LocalGroup),
		"Member": fiber.Locals[database.Member](ctx, LocalMember),
		"Rights": fiber.Locals[database.Role](ctx, LocalRights),
	}

	bindx = MapMerge(&bindx, bind)
	return bindx
}

// Converts the pointer to a value
func MapMerge(maps ...*fiber.Map) fiber.Map {
	merge := fiber.Map{}
	for _, m := range maps {
		if m == nil {
			continue
		}

		for k, v := range *m {
			merge[k] = v
		}
	}

	return merge
}

func RenderBuffer(app *fiber.App, templateName string, bind any) (bytes.Buffer, error) {
	buf := bytes.NewBuffer([]byte{})
	err := app.Config().Views.Render(buf, templateName, bind)
	if err != nil {
		log.Error(err)
		buf.WriteString(template.HTMLEscapeString(err.Error()))
	}

	return *buf, err
}

func WrapOob(swap string, message *string) string {
	msg := ""
	if message != nil {
		msg = *message
	}

	return "<div hx-swap-oob=\"" + swap + "\">" + msg + "</div>"
}
