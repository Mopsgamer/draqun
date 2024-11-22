package internal

import (
	"restapp/internal/environment"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Initialize gofiber application, including DB and view engine.
func NewApp() (*fiber.App, error) {
	environment.WaitForBuild()

	db, err := InitDB()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	UseResponder := func(handler func(r Responder) error) fiber.Handler {
		return func(c fiber.Ctx) error {
			return handler(Responder{c, *db})
		}
	}

	UsePage := func(guestRedirect, templatePath, title string, bind *fiber.Map, layouts ...string) fiber.Handler {
		bindx := fiber.Map{
			"Title": "Restapp - " + title,
		}
		if bind != nil {
			for k, v := range *bind {
				bindx[k] = v
			}
		}
		return UseResponder(func(r Responder) error {
			return r.RenderPage(
				guestRedirect,
				templatePath,
				bindx,
				layouts...,
			)
		})
	}

	app := fiber.New(fiber.Config{
		Views:             NewAppHtmlEngine(db),
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	// static
	app.Get("/static/*", static.New("./web/static", static.Config{Browse: true}))
	app.Get("/assets/*", static.New("./web/assets", static.Config{Browse: true}))
	app.Get("/partials", static.New("./web/templates/partials", static.Config{Browse: true}))
	app.Get("/partials/*", UseResponder(func(r Responder) error { return r.RenderTemplate() }))

	// get
	app.Get("/", UsePage("", "index", "Home page", nil, "partials/main"))
	app.Get("/settings", UsePage("/", "settings", "Settings", nil, "partials/main"))
	renderChat := UsePage("", "chat", "Settings", &fiber.Map{"IsChatPage": true})
	app.Get("/chat", renderChat)
	app.Get("/chat/groups/:group_id", renderChat)

	// post
	app.Post("/account/create", UseResponder(func(r Responder) error { return r.UserSignUp() }))
	app.Post("/account/login", UseResponder(func(r Responder) error { return r.UserLogin() }))
	app.Post("/groups/create", UseResponder(func(r Responder) error { return r.GroupCreate() }))

	// put
	app.Put("/account/change/name", UseResponder(func(r Responder) error { return r.UserChangeName() }))
	app.Put("/account/change/email", UseResponder(func(r Responder) error { return r.UserChangeEmail() }))
	app.Put("/account/change/phone", UseResponder(func(r Responder) error { return r.UserChangePhone() }))
	app.Put("/account/change/password", UseResponder(func(r Responder) error { return r.UserChangePassword() }))
	app.Put("/account/logout", UseResponder(func(r Responder) error { return r.UserLogout() }))
	// TODO: app.Put("/groups/:group_id/change", UseResponder(func(r Responder) error { return r.GroupChange() }))

	// delete
	app.Delete("/groups/:group_id/leave", UseResponder(func(r Responder) error { return r.GroupLeave() }))
	app.Delete("/groups/:group_id", UseResponder(func(r Responder) error { return r.GroupDelete() }))
	app.Delete("/account/delete", UseResponder(func(r Responder) error { return r.UserDelete() }))

	// websoket
	// https://docs.gofiber.io/contrib/next/websocket/
	// TODO: ws - update messages
	// TODO: ws - update members

	app.Use(UsePage("", "partials/x", strconv.Itoa(fiber.StatusNotFound), &fiber.Map{
		"StatusCode":    fiber.StatusNotFound,
		"StatusMessage": fiber.ErrNotFound.Message,
	}, "partials/main"))

	return app, nil
}
