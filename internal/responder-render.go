package internal

import (
	"restapp/internal/model"
	"restapp/internal/model_request"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

// Render a page using a template.
func (r Responder) RenderPage(guestRedirect string, templatePath string, bind fiber.Map, layouts ...string) error {
	user := r.User()
	if guestRedirect != "" && user == nil {
		return r.Redirect().To(guestRedirect)
	}
	return r.Render(templatePath, r.PageMap(bind), layouts...)
}

func (r *Responder) PageMap(bind fiber.Map) fiber.Map {
	user := r.User()
	result := fiber.Map{}
	if user != nil {
		result["User"] = user
	} else {
		result["TokenError"] = true
		result["Message"] = "Authorization error"
		result["Id"] = "local-token-error"
	}

	groupUri := new(model_request.GroupUri)
	if err := r.Bind().URI(groupUri); err != nil {
		log.Error(err)
	} else if groupUri.GroupId != nil {
		group := r.DB.GroupById(*groupUri.GroupId)
		result["Group"] = group
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
	Id           string     `query:"id"`
	Message      string     `query:"message"`
	OpenSettings bool       `query:"open-settings"`
	OpenSignUp   bool       `query:"open-signup"`
	OpenLogin    bool       `query:"open-login"`
	User         model.User `query:"user"` // its safe
}

// Renders a template, requested by a client.
func (r Responder) RenderTemplate() error {
	q := new(HTMXPartialQuery)
	err := r.Bind().Query(q)
	r.User()
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
