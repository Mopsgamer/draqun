package internal

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// Render a page using a template.
func (r Responder) RenderPage(templatePath string, bind fiber.Map, layouts ...string) error {
	return r.Render(templatePath, r.PageMap(bind), layouts...)
}

func (r *Responder) PageMap(bind fiber.Map) fiber.Map {
	user, errToken := r.GetOwner()
	if errToken != nil {
		log.Error(errToken)
		r.Cookie(&fiber.Cookie{
			Name:    "Authorization",
			Value:   "",
			Expires: time.Now(),
		})
	}
	result := fiber.Map{}
	if errToken != nil {
		result["TokenError"] = true
		result["Message"] = "Authorization error"
		result["Id"] = "local-token-error"
	} else {
		result["User"] = user
	}

	for k, v := range bind {
		result[k] = v
	}
	return result
}

// This type describes ALL values in EVERY partial, which can be passed into ./templates/partials
// and used by htmx requests to replace DOM, using template generation through get requests
//
// EXAMPLE:
//
//	<div hx-get="/partials/chat?mode=compact">
//
// NOTE: wont move this to internal/htmx.go
// since its only for the RenderTemplate
type HTMXPartialQuery struct {
	Id           string `query:"id"`
	Message      string `query:"message"`
	OpenSettings bool   `query:"open-settings"`
	OpenRegister bool   `query:"open-register"`
	OpenLogin    bool   `query:"open-login"`
	User         User   `query:"user"` // its safe
}

// Renders a template, requested by a client.
func (r Responder) RenderTemplate() error {
	q := new(HTMXPartialQuery)
	err := r.Bind().Query(q)
	r.GetOwner()
	if err != nil {
		return err
	}
	return r.Render(r.Path()[1:], r.PageMap(fiber.Map{}))
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
