package logic

import (
	"bytes"
	"html/template"
	"restapp/internal/logic/database"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

type Logic struct {
	DB *database.Database
}

// Converts the pointer to the value
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
