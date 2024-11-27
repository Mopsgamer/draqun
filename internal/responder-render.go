package internal

import (
	"bytes"

	"github.com/gofiber/fiber/v3"
)

// Should return redirect path or empty string.
type RedirectLogic func(r Responder, bind *fiber.Map) string

func (r Responder) RenderBuffer(template string, bind any) *bytes.Buffer {
	buf := bytes.NewBuffer([]byte{})
	r.Ctx.App().Config().Views.Render(buf, template, bind)

	return buf
}

// Render a page using a template.
// Special
func (r Responder) RenderPage(templatePath string, bind *fiber.Map, redirect RedirectLogic, layouts ...string) error {
	bindx := r.MapPage(bind)
	if redirect != nil {
		if path := redirect(r, bind); path != "" {
			return r.Ctx.Redirect().To(path)
		}
	}
	if title, ok := (*bind)["Title"].(string); ok {
		bindx["Title"] = "Restapp - " + title
	}
	return r.Ctx.Render(templatePath, bindx, layouts...)
}

func (r Responder) MapPage(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{}
	if user, tokenErr := r.User(); user != nil {
		bindx["User"] = user
	} else if tokenErr != nil {
		bindx["TokenError"] = true
		bindx["Message"] = "Authorization error"
		bindx["Id"] = "local-token-error"
	}

	if group := r.Group(); group != nil {
		bindx["Group"] = group
	}

	bindx = r.Map(bind)
	return bindx
}

// Renders the danger message html element.
func (r Responder) RenderDanger(message, id string) error {
	return r.Ctx.Render("partials/danger", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the warning message html element.
func (r Responder) RenderWarning(message, id string) error {
	return r.Ctx.Render("partials/warning", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the success message html element.
func (r Responder) RenderSuccess(message, id string) error {
	return r.Ctx.Render("partials/success", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}
