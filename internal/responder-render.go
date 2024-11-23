package internal

import (
	"github.com/gofiber/fiber/v3"
)

// Should return redirect path or empty string.
type RedirectLogic func(r Responder, bind *fiber.Map) string

// Render a page using a template.
// Special
func (r Responder) RenderPage(templatePath string, bind *fiber.Map, redirect RedirectLogic, layouts ...string) error {
	bindx := r.PageMap(bind)
	if redirect != nil {
		if path := redirect(r, bind); path != "" {
			return r.Redirect().To(path)
		}
	}
	if title, ok := (*bind)["Title"].(string); ok {
		bindx["Title"] = "Restapp - " + title
	}
	return r.Render(templatePath, bindx, layouts...)
}

func (r *Responder) PageMap(bind *fiber.Map) fiber.Map {
	result := fiber.Map{}
	if user, tokenErr := r.User(); user != nil {
		result["User"] = user
	} else if tokenErr != nil {
		result["TokenError"] = true
		result["Message"] = "Authorization error"
		result["Id"] = "local-token-error"
	}

	if group := r.Group(); group != nil {
		result["Group"] = group
	}

	if bind != nil {
		for k, v := range *bind {
			result[k] = v
		}
	}
	return result
}

// Renders the danger message html element.
func (r Responder) RenderDanger(message, id string) error {
	return r.Render("partials/danger", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the warning message html element.
func (r Responder) RenderWarning(message, id string) error {
	return r.Render("partials/warning", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the success message html element.
func (r Responder) RenderSuccess(message, id string) error {
	return r.Render("partials/success", fiber.Map{
		"Id":      id,
		"Message": message,
	})
}
