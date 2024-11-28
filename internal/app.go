package internal

import (
	"errors"
	"restapp/internal/environment"
	i18n "restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/database"
	"restapp/internal/logic/logic_http"
	"restapp/internal/logic/logic_websocket"
	"restapp/websocket"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

// Initialize gofiber application, including DB and view engine.
func NewApp() (*fiber.App, error) {
	environment.WaitForBuild()

	db, err := database.InitDB()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	UseHTTP := func(handler func(r logic_http.LogicHTTP) error) fiber.Handler {
		return func(c fiber.Ctx) error {
			logic := logic_http.LogicHTTP{
				Logic: logic.Logic{Ctx: c, DB: db},
				Ctx:   c,
			}
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return handler(logic)
		}
	}

	UseWebsocket := func(acceptWrite func(r logic_websocket.LogicWebsocket, template string, bind any) (bool, error), handlerRead func(r logic_websocket.LogicWebsocket, messageType int, message []byte) error) fiber.Handler {
		return func(httpc fiber.Ctx) error {
			http := logic_http.LogicHTTP{
				Logic: logic.Logic{Ctx: httpc, DB: db},
				Ctx:   httpc,
			}
			if websocket.IsWebSocketUpgrade(httpc) {
				user, _ := http.User()
				if user == nil {
					return http.RenderWarning(i18n.MessageErrUserNotFound, "ws-error")
				}

				return websocket.New(func(wsc *websocket.Conn) {
					ws := logic_websocket.LogicWebsocket{
						Logic:  logic.Logic{Ctx: wsc, DB: db},
						Closed: false,
						Ctx:    *wsc,
						Accept: acceptWrite,
					}
					logic_websocket.WebsocketConnections.UserConnect(user.Id, &ws)
					defer logic_websocket.WebsocketConnections.UserDisconnect(user.Id, &ws)
					for !ws.Closed {
						messageType, message, err := ws.Ctx.ReadMessage()
						if err != nil {
							log.Info(err)
							break
						}

						err = handlerRead(ws, messageType, message)
						if err != nil {
							log.Error(err)
							break
						}
					}
				})(httpc)
			}
			return httpc.Next()
		}
	}

	UsePage := func(templatePath string, bind *fiber.Map, redirectOn logic_http.RedirectCompute, layouts ...string) fiber.Handler {
		bindx := fiber.Map{
			"Title": "?",
		}
		if bind != nil {
			for k, v := range *bind {
				bindx[k] = v
			}
		}
		return UseHTTP(func(r logic_http.LogicHTTP) error {
			return r.RenderPage(
				templatePath,
				&bindx,
				redirectOn,
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
	app.Get("/", UsePage("index", &fiber.Map{"Title": "Discover"}, func(r logic_http.LogicHTTP, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/settings", UsePage("settings", &fiber.Map{"Title": "Settings"},
		func(r logic_http.LogicHTTP, bind *fiber.Map) string {
			if user, _ := r.User(); user == nil {
				return "/"
			}
			return ""
		}, "partials/main"))
	app.Get("/chat", UsePage("chat", &fiber.Map{"Title": "Home", "IsChatPage": true},
		func(r logic_http.LogicHTTP, bind *fiber.Map) string {
			return ""
		}))
	app.Get("/chat/groups/:group_id", UsePage("chat", &fiber.Map{"Title": "Group", "IsChatPage": true},
		func(r logic_http.LogicHTTP, bind *fiber.Map) string {
			group := r.Group()
			if group == nil {
				return "/chat"
			}

			(*bind)["Title"] = group.Nick
			return ""
		}))

	// post
	app.Post("/account/create", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserSignUp() }))
	app.Post("/account/login", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserLogin() }))
	app.Post("/groups/create", UseHTTP(func(r logic_http.LogicHTTP) error { return r.GroupCreate() }))
	app.Post("/groups/:group_id/messages/create", UseHTTP(func(r logic_http.LogicHTTP) error { return r.MessageCreate() }))

	// put
	app.Put("/account/change/name", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserChangeName() }))
	app.Put("/account/change/email", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserChangeEmail() }))
	app.Put("/account/change/phone", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserChangePhone() }))
	app.Put("/account/change/password", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserChangePassword() }))
	app.Put("/account/logout", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserLogout() }))
	// TODO: app.Put("/groups/:group_id/change", UseHTTP(func(r logic_http.LogicHTTP) error { return r.GroupChange() }))

	// delete
	app.Delete("/groups/:group_id/leave", UseHTTP(func(r logic_http.LogicHTTP) error { return r.GroupLeave() }))
	app.Delete("/groups/:group_id", UseHTTP(func(r logic_http.LogicHTTP) error { return r.GroupDelete() }))
	app.Delete("/account/delete", UseHTTP(func(r logic_http.LogicHTTP) error { return r.UserDelete() }))

	// websoket
	app.Get("/groups/:group_id/messages", UseWebsocket(
		func(r logic_websocket.LogicWebsocket, template string, bind any) (bool, error) {
			group := r.Group()
			if group == nil {
				return false, errors.New("group " + r.Ctx.Params("group_id") + " not found")
			}

			if template != "partials/message" {
				return false, nil
			}

			// HACK: user should be a member and have read permissions
			return true, nil
		},
		func(r logic_websocket.LogicWebsocket, messageType int, message []byte) error {
			return nil
		},
	))
	app.Get("/groups/:group_id/users", UseWebsocket(
		func(r logic_websocket.LogicWebsocket, template string, bind any) (bool, error) {
			group := r.Group()
			if group == nil {
				return false, errors.New("group " + r.Ctx.Params("group_id") + " not found")
			}

			// HACK: user should be a member and have read permissions
			if template != "partials/group-member" {
				return false, nil
			}

			return true, nil
		},
		func(r logic_websocket.LogicWebsocket, messageType int, message []byte) error {
			return nil
		},
	))

	app.Use(UsePage("partials/x", &fiber.Map{
		"Title":         strconv.Itoa(fiber.StatusNotFound),
		"StatusCode":    fiber.StatusNotFound,
		"StatusMessage": fiber.ErrNotFound.Message,
	}, func(r logic_http.LogicHTTP, bind *fiber.Map) string { return "" }, "partials/main"))

	return app, nil
}
