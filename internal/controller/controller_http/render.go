package controller_http

import (
	"restapp/internal/controller"
	"restapp/internal/environment"
	"restapp/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

// Should return redirect path or empty string.
type RedirectCompute func(r ControllerHttp, bind *fiber.Map) string

// Render a page using a template.
// Special
func (r ControllerHttp) RenderPage(templatePath string, bind *fiber.Map, redirect RedirectCompute, layouts ...string) error {
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

func (r ControllerHttp) MapPage(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{}

	bindx["DenoJson"] = environment.DenoJson
	bindx["GoMod"] = environment.GoMod

	rights, member, user, group := r.Rights()
	if user != nil {
		bindx["User"] = user
	}

	if group != nil {
		bindx["Group"] = group
	}

	if member != nil {
		bindx["Member"] = member
		bindx["Rights"] = rights
	}

	bindx = controller.MapMerge(&bindx, bind)
	return bindx
}

func (r ControllerHttp) RenderString(template string, bind any) (string, error) {
	return controller.RenderString(r.Ctx.App(), template, bind)
}

func wrapRenderNotice(r ControllerHttp, template, message, id string) error {
	return r.Ctx.Render(template, fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the danger message html element.
func (r ControllerHttp) RenderInternalError(id string) error {
	r.Ctx.Status(fiber.StatusInternalServerError)
	return r.RenderDanger(i18n.MessageErrInvalidRequest, id)
}

// Renders the danger message html element.
func (r ControllerHttp) RenderDanger(message, id string) error {
	return wrapRenderNotice(r, "partials/danger", message, id)
}

// Renders the warning message html element.
func (r ControllerHttp) RenderWarning(message, id string) error {
	return wrapRenderNotice(r, "partials/warning", message, id)
}

// Renders the success message html element.
func (r ControllerHttp) RenderSuccess(message, id string) error {
	return wrapRenderNotice(r, "partials/success", message, id)
}
