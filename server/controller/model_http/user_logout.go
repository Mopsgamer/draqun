package model_http

import (
	"time"

	"github.com/Mopsgamer/draqun/server/controller/controller_http"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
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

	log.Info("logout and redirect ", ctl.HTMXCurrentPath())
	ctl.HTMXRedirect(ctl.HTMXCurrentPath())
	return ctl.Ctx.Render("partials/redirecting", fiber.Map{})
}
