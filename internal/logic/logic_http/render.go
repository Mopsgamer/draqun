package logic_http

import (
	"restapp/internal/logic"

	"github.com/gofiber/fiber/v3"
)

// Should return redirect path or empty string.
type RedirectCompute func(r LogicHTTP, bind *fiber.Map) string

// Render a page using a template.
// Special
func (r LogicHTTP) RenderPage(templatePath string, bind *fiber.Map, redirect RedirectCompute, layouts ...string) error {
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

func (r LogicHTTP) MapPage(bind *fiber.Map) fiber.Map {
	bindx := fiber.Map{}

	rights, member, user, group := r.Rights()
	if user != nil {
		bindx["User"] = user
	}

	if group != nil {
		bindx["Group"] = group
		bindx["GroupLink"] = r.Ctx.Host() + logic.PathRedirectGroupJoin(group.Name)
	}

	if member != nil {
		bindx["Member"] = member
		bindx["Rights"] = rights
	}

	bindx = logic.MapMerge(&bindx, bind)
	return bindx
}

func (r LogicHTTP) RenderString(template string, bind any) (string, error) {
	return logic.RenderString(r.Ctx.App(), template, bind)
}

func wrapRenderNotice(r LogicHTTP, template, message, id string) error {
	return r.Ctx.Render(template, fiber.Map{
		"Id":      id,
		"Message": message,
	})
}

// Renders the danger message html element.
func (r LogicHTTP) RenderDanger(message, id string) error {
	return wrapRenderNotice(r, "partials/danger", message, id)
}

// Renders the warning message html element.
func (r LogicHTTP) RenderWarning(message, id string) error {
	return wrapRenderNotice(r, "partials/warning", message, id)
}

// Renders the success message html element.
func (r LogicHTTP) RenderSuccess(message, id string) error {
	return wrapRenderNotice(r, "partials/success", message, id)
}
