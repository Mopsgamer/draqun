package internal

import (
	"errors"
	"restapp/internal/environment"
	"restapp/internal/logic"
	"restapp/internal/logic/database"
	"restapp/internal/logic/logic_http"
	"restapp/internal/logic/logic_websocket"
	"restapp/internal/logic/model"
	"restapp/internal/logic/model_request"
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

	UseWebsocket := func(acceptWrite func(r logic_http.LogicHTTP) error, updateContent func(r *logic_websocket.LogicWebsocket) *string, handler func(r logic_websocket.LogicWebsocket) error) fiber.Handler {
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

			user, _ := http.User()
			group := http.Group()

			websocket.New(func(conn *websocket.Conn) {
				// NOTE: Inside this method the 'c' variable is already updated
				// and some methods can not work. They should be used outside.
				ws := logic_websocket.New(
					appLogic,
					conn,
					app,
					updateContent,
					&fiber.Map{
						"User":  user,
						"Group": group,
					},
				)

				logic_websocket.WebsocketConnections.UserConnect(user.Id, &ws)
				defer logic_websocket.WebsocketConnections.UserDisconnect(user.Id, &ws)
				for !ws.Closed {
					err = handler(ws)
					if err != nil {
						break
					}
				}
				ws.Ctx.Close()
			})(http.Ctx)

			return nil
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

			// HACK: user should be a member and have read permissions
			return nil
		},
		func(ws *logic_websocket.LogicWebsocket) *string {
			str := logic.RenderString(app, "partials/chat-group", ws.Bind)
			return str
		},
		func(ws logic_websocket.LogicWebsocket) error {
			group, _ := (*ws.Bind)["Group"].(*model.Group)
			user, _ := (*ws.Bind)["User"].(*model.User)
			messageCreate := new(model_request.MessageCreate)
			err = ws.Ctx.ReadJSON(messageCreate)
			if err != nil {
				log.Error(err)
			}
			messageCreate.GroupId = group.Id
			message := messageCreate.Message(user.Id)
			if !model.IsValidMessageContent(messageCreate.Content) {
				ws.WebsocketRender("partials/warning", fiber.Map{"Message": "Invalid message content"})
				return nil
			}

			// TODO: ws: add validation for message and move in separate method

			ws.MessageSend(*message)
			err = ws.SendContent()
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
