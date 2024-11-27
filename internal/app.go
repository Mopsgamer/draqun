package internal

import (
	"errors"
	"restapp/internal/environment"
	"restapp/websocket"
	"slices"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

var (
	// A websocket connection list for each user id.
	WebsocketConnections = map[uint64]([]*ResponderWebsocket){}
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
			responder := Responder{c, *db}
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return handler(responder)
		}
	}

	UseWebsocket := func(acceptWrite func(r ResponderWebsocket, template string, bind any) (bool, error), handlerRead func(r ResponderWebsocket, messageType int, message []byte) error) fiber.Handler {
		return func(c fiber.Ctx) error {
			responder := Responder{c, *db}
			if websocket.IsWebSocketUpgrade(c) {
				user, err := responder.User()
				if user == nil {
					return err
				}

				return websocket.New(func(c *websocket.Conn) {
					responderWS := &ResponderWebsocket{
						Responder: responder,
						WS:        *c,
						Accept:    acceptWrite,
					}
					s := WebsocketConnections[user.Id]
					s = append(s, responderWS)
					defer func() {
						i := slices.Index(s, responderWS)
						WebsocketConnections[user.Id] = slices.Delete(s, i, i+1)
					}()
					for {
						messageType, message, err := responderWS.WS.ReadMessage()
						if err != nil {
							log.Error(err)
							break
						}

						err = handlerRead(*responderWS, messageType, message)
						if err != nil {
							log.Error(err)
							break
						}
					}
				})(c)
			}
			return c.Next()
		}
	}

	UsePage := func(templatePath string, bind *fiber.Map, redirectLogic RedirectLogic, layouts ...string) fiber.Handler {
		bindx := fiber.Map{
			"Title": "?",
		}
		if bind != nil {
			for k, v := range *bind {
				bindx[k] = v
			}
		}
		return UseResponder(func(r Responder) error {
			return r.RenderPage(
				templatePath,
				&bindx,
				redirectLogic,
				layouts...,
			)
		})
	}

	engine := NewAppHtmlEngine(db)
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	// static
	app.Get("/static/*", static.New("./web/static", static.Config{Browse: true}))
	app.Get("/partials*", static.New("./web/templates/partials", static.Config{Browse: true}))

	// get
	app.Get("/", UsePage("index", &fiber.Map{"Title": "Discover"}, func(r Responder, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/settings", UsePage("settings", &fiber.Map{"Title": "Settings"},
		func(r Responder, bind *fiber.Map) string {
			if user, _ := r.User(); user == nil {
				return "/"
			}
			return ""
		}, "partials/main"))
	app.Get("/chat", UsePage("chat", &fiber.Map{"Title": "Home", "IsChatPage": true},
		func(r Responder, bind *fiber.Map) string {
			return ""
		}))
	app.Get("/chat/groups/:group_id", UsePage("chat", &fiber.Map{"Title": "Group", "IsChatPage": true},
		func(r Responder, bind *fiber.Map) string {
			group := r.Group()
			if group == nil {
				return "/chat"
			}

			(*bind)["Title"] = group.Nick
			return ""
		}))

	// post
	app.Post("/account/create", UseResponder(func(r Responder) error { return r.UserSignUp() }))
	app.Post("/account/login", UseResponder(func(r Responder) error { return r.UserLogin() }))
	app.Post("/groups/create", UseResponder(func(r Responder) error { return r.GroupCreate() }))
	app.Post("/groups/:group_id/messages/create", UseResponder(func(r Responder) error { return r.MessageCreate() }))

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
	app.Get("/groups/:group_id/messages", UseWebsocket(
		func(r ResponderWebsocket, template string, bind any) (bool, error) {
			group := r.Group()
			if group == nil {
				return false, errors.New("group " + r.Ctx.Params("group_id") + " not found")
			}

			if template != "partials/message" {
				return false, nil
			}

			return true, nil
		},
		func(r ResponderWebsocket, messageType int, message []byte) error {
			return nil
		},
	))
	app.Get("/groups/:group_id/users", UseWebsocket(
		func(r ResponderWebsocket, template string, bind any) (bool, error) {
			group := r.Group()
			if group == nil {
				return false, errors.New("group " + r.Ctx.Params("group_id") + " not found")
			}

			if template != "partials/group-member" {
				return false, nil
			}

			return true, nil
		},
		func(r ResponderWebsocket, messageType int, message []byte) error {
			return nil
		},
	))

	app.Use(UsePage("partials/x", &fiber.Map{
		"Title":         strconv.Itoa(fiber.StatusNotFound),
		"StatusCode":    fiber.StatusNotFound,
		"StatusMessage": fiber.ErrNotFound.Message,
	}, func(r Responder, bind *fiber.Map) string { return "" }, "partials/main"))

	return app, nil
}
