package controller_http

import (
	"restapp/internal/controller"
	"restapp/internal/environment"
	"restapp/internal/i18n"

	"github.com/gofiber/fiber/v3"
)

// Should return redirect path or empty string.
type RedirectCompute func(ctl ControllerHttp, bind *fiber.Map) string

// Render a page using a template.
// Special
func (r ControllerHttp) RenderPage(templatePath string, bind *fiber.Map, redirect RedirectCompute, layouts ...string) error {
	bindx := r.MapPage(bind)
	if path := redirect(r, bind); path != "" {
		return r.Ctx.Redirect().To(path)
	}
	if title, ok := (*bind)["Title"].(string); ok {
		bindx["Title"] = "Restapp - " + title
	}
	return r.Ctx.Render(templatePath, bindx, layouts...)
}

func (ctl ControllerHttp) MapPage(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{}

	bindx["DenoJson"] = environment.DenoJson
	bindx["GoMod"] = environment.GoMod
	bindx["GitHash"] = environment.GitHash
	bindx["GitHashLong"] = environment.GitHashLong

	bindx = controller.MapMerge(&bindx, bind)
	return bindx
}

func (ctl ControllerHttp) RenderString(template string, bind any) (string, error) {
	return controller.RenderString(ctl.Ctx.App(), template, bind)
}

func wrapRenderNotice(r ControllerHttp, template, message, id string) error {
	return r.Ctx.Render(template, fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the danger message html element.
func (ctl ControllerHttp) RenderInternalError(id string) error {
	ctl.Ctx.Status(fiber.StatusInternalServerError)
	return ctl.RenderDanger(i18n.MessageErrInvalidRequest, id)
}

// Renders the danger message html element.
func (ctl ControllerHttp) RenderDanger(message, id string) error {
	return wrapRenderNotice(ctl, "partials/danger", message, id)
}

// Renders the warning message html element.
func (ctl ControllerHttp) RenderWarning(message, id string) error {
	return wrapRenderNotice(ctl, "partials/warning", message, id)
}

// Renders the success message html element.
func (ctl ControllerHttp) RenderSuccess(message, id string) error {
	return wrapRenderNotice(ctl, "partials/success", message, id)
}
