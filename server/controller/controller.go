package controller

import (
	"bytes"
	"html/template"

	"github.com/Mopsgamer/draqun/server/controller/database"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

type Controller struct {
	Ctx fiber.Ctx
	DB  database.Database
}

type Handler func(ctl Controller) error

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

func RenderString(app *fiber.App, template string, bind any) (string, error) {
	buf, err := RenderBuffer(app, template, bind)

	str := buf.String()
	return str, err
}

func WrapOob(swap string, message *string) string {
	msg := ""
	if message != nil {
		msg = *message
	}

	return "<div hx-swap-oob=\"" + swap + "\">" + msg + "</div>"
}
