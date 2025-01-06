package model_http

import (
	"restapp/internal/controller/controller_http"
	"time"

	"github.com/gofiber/fiber/v3"
)

type UserLogout struct {
	CookieUserToken
}

func (request *UserLogout) HandleHtmx(ctl controller_http.ControllerHttp) error {
	ctl.Ctx.Cookie(&fiber.Cookie{
		Name:    "Authorization",
		Value:   "",
		Expires: time.Now(),
	})

	ctl.HTMXRedirect(ctl.HTMXCurrentPath())
	return ctl.Ctx.Render("partials/redirecting", fiber.Map{})
}
