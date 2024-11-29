package internal

import (
	"errors"
	"fmt"
	"restapp/internal/environment"
	"restapp/internal/i18n"
	"restapp/internal/logic"
	"restapp/internal/logic/database"
	"restapp/internal/logic/logic_http"
	"restapp/internal/logic/logic_websocket"
	"restapp/internal/logic/model_request"
	"restapp/websocket"
	"strconv"
	"time"

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

	engine := NewAppHtmlEngine(db)
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	appLogic := &logic.Logic{DB: db}
	UseHTTP := func(handler func(r logic_http.LogicHTTP) error) fiber.Handler {
		return func(c fiber.Ctx) error {
			logic := logic_http.LogicHTTP{
				Logic: appLogic,
				Ctx:   c,
			}
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return handler(logic)
		}
	}

	UseWebsocket := func(
		acceptWrite func(r logic_http.LogicHTTP) error,
		handler func(r *logic_websocket.LogicWebsocket,
			wsmsg *model_request.WebsocketMessage) error,
	) fiber.Handler {
		return func(c fiber.Ctx) error {
			http := logic_http.LogicHTTP{
				Logic: appLogic,
				Ctx:   c,
			}
			if !websocket.IsWebSocketUpgrade(http.Ctx) {
				http.Ctx.Next()
			}

			err := acceptWrite(http)

			if err != nil {
				log.Error(err)
				message := err.Error()
				return http.RenderDanger(message, "ws-err")
			}

			ip := http.Ctx.IP()
			user, _ := http.User()
			group := http.Group()

			websocket.New(func(conn *websocket.Conn) {
				// NOTE: Inside this method the 'c' variable is already updated
				// and some methods can not work. They should be used outside.
				ws := logic_websocket.New(
					appLogic,
					conn,
					app,
					ip,
					&fiber.Map{
						"User":  user,
						"Group": group,
					},
				)

				logic_websocket.WebsocketConnections.UserConnect(user.Id, &ws)
				defer logic_websocket.WebsocketConnections.UserDisconnect(user.Id, &ws)
				for !ws.Closed {
					messageType, message, err := ws.Ctx.ReadMessage()
					if err != nil {
						log.Error(err)
						break
					}

					start := time.Now()
					ws.MessageType = messageType
					ws.Message = message
					wsmsg := new(model_request.WebsocketMessage)
					if ws.MessageType == websocket.TextMessage {
						parseErr := ws.GetMessageJSON(wsmsg)
						if parseErr != nil {
							log.Error(parseErr)
						}
					} else {
						wsmsg = nil
					}
					err = handler(&ws, wsmsg)

					colorErr := fiber.DefaultColors.Green
					if err != nil {
						colorErr = fiber.DefaultColors.Red
					}

					t := "?"
					if wsmsg != nil {
						t = wsmsg.Type
					}

					fmt.Printf(
						"%s | %s%3s%s | %13s | %15s | %s%s%s | %d | %s | %s\"%s\"%s\n",
						time.Now().Format("15:04:05"),
						colorErr,
						"ws",
						fiber.DefaultColors.Reset,
						time.Since(start),
						ws.IP,
						fiber.DefaultColors.Cyan,
						t,
						fiber.DefaultColors.Reset,
						ws.MessageType,
						ws.Message,
						colorErr,
						*logic_websocket.LastWrite,
						fiber.DefaultColors.Reset,
					)
					if err != nil {
						break
					}
				}
				ws.Ctx.Close()
			})(http.Ctx)

			return nil
		}
	}

	UsePage := func(
		templatePath string,
		bind *fiber.Map,
		redirectOn logic_http.RedirectCompute,
		layouts ...string,
	) fiber.Handler {
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
	app.Get("/chat/groups/:group_id", UseWebsocket(
		func(ws logic_http.LogicHTTP) error {
			user, err := ws.User()
			if user == nil {
				return err
			}

			group := ws.Group()
			if group == nil {
				return errors.New("group " + ws.Ctx.Params("group_id") + " not found")
			}

			// FIXME: user should be a member and have read permissions
			return nil
		},
		func(ws *logic_websocket.LogicWebsocket, wsmsg *model_request.WebsocketMessage) error {
			if wsmsg == nil {
				ws.SendWarning(i18n.MessageErrInvalidRequest, "ws-error")
				return nil
			}

			if wsmsg.Type == "ping" {
				err = ws.SendString("")
			} else if wsmsg.Type == "send-message-form" {
				err = ws.MessageCreate()
			}

			if err != nil {
				log.Error(err)
			}
			return err
		},
	))

	app.Use(UsePage("partials/x", &fiber.Map{
		"Title":         strconv.Itoa(fiber.StatusNotFound),
		"StatusCode":    fiber.StatusNotFound,
		"StatusMessage": fiber.ErrNotFound.Message,
	}, func(r logic_http.LogicHTTP, bind *fiber.Map) string { return "" }, "partials/main"))

	return app, nil
}
