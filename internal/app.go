package internal

import (
	"fmt"
	"reflect"
	"restapp/internal/controller"
	"restapp/internal/controller/controller_http"
	"restapp/internal/controller/controller_ws"
	"restapp/internal/controller/database"
	"restapp/internal/controller/model_http"
	"restapp/internal/environment"
	"restapp/websocket"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/graphql-go/graphql"
)

// Initialize gofiber application, including DB and view engine.
func NewApp() (*fiber.App, error) {
	environment.WaitForBuild()

	db, errDBLoad := database.InitDB()
	if errDBLoad != nil {
		log.Error(errDBLoad)
		return nil, errDBLoad
	}

	schema, graphqlFields, errGraphqlLoad := initGraphql(*db)
	if errGraphqlLoad != nil {
		log.Error(errGraphqlLoad)
		return nil, errGraphqlLoad
	}

	engine := NewAppHtmlEngine(db)
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	ctl := &controller.Controller{DB: db}
	UseHTTP := func(handler func(r controller_http.ControllerHttp) error) fiber.Handler {
		return func(c fiber.Ctx) error {
			controller := controller_http.ControllerHttp{
				Controller: ctl,
				Ctx:        c,
			}
			if websocket.IsWebSocketUpgrade(c) {
				return c.Next()
			}
			return handler(controller)
		}
	}

	UseHTTPPage := func(
		templatePath string,
		bind *fiber.Map,
		redirectOn controller_http.RedirectCompute,
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
		return UseHTTP(func(r controller_http.ControllerHttp) error {
			return r.RenderPage(
				templatePath,
				&bindx,
				redirectOn,
				layouts...,
			)
		})
	}

	UseWebsocket := func(
		subscribe []string,
		handler func(ws *controller_ws.ControllerWs) error,
	) fiber.Handler {
		return func(c fiber.Ctx) error {
			ctlHttp := controller_http.ControllerHttp{
				Controller: ctl,
				Ctx:        c,
			}
			if !websocket.IsWebSocketUpgrade(ctlHttp.Ctx) {
				ctlHttp.Ctx.Next()
			}

			ip := ctlHttp.Ctx.IP()
			bind := ctlHttp.MapPage(nil)

			websocket.New(func(conn *websocket.Conn) {
				// NOTE: Inside this method the 'c' variable is already updated
				// and some methods can not work. They should be used outside.
				ctlWs := controller_ws.New(
					ctl,
					conn,
					app,
					ip,
					&bind,
				)

				user := ctlWs.User()

				controller_ws.UserSessionMap.Connect(user.Id, &ctlWs)
				defer controller_ws.UserSessionMap.Close(user.Id, &ctlWs)
				for !ctlWs.Closed {
					messageType, message, err := ctlWs.Ctx.ReadMessage()
					if err != nil {
						break
					}

					// start := time.Now()
					ctlWs.MessageType = messageType
					ctlWs.Message = message
					err = handler(&ctlWs)

					// colorErr := fiber.DefaultColors.Green
					// if err != nil {
					// 	colorErr = fiber.DefaultColors.Red
					// }

					// fmt.Printf(
					// 	"%s | %s%3s%s | %13s | %15s | %d | %s%s%s \n",
					// 	time.Now().Format("15:04:05"),
					// 	colorErr,
					// 	"ws",
					// 	fiber.DefaultColors.Reset,
					// 	time.Since(start),
					// 	ws.IP,
					// 	ws.MessageType,
					// 	fiber.DefaultColors.Yellow,
					// 	ws.Message,
					// 	fiber.DefaultColors.Reset,
					// )
					if err != nil {
						break
					}
				}
				ctlWs.Ctx.Close()
			})(ctlHttp.Ctx)

			return nil
		}
	}

	// static
	app.Get("/static/*", static.New("./web/static", static.Config{Browse: true}))
	app.Get("/partials*", static.New("./web/templates/partials", static.Config{Browse: true}))

	// pages
	docs := initDocs()
	app.Get("/", UseHTTPPage("index", &fiber.Map{"Title": "Discover", "IsHomePage": true}, func(r controller_http.ControllerHttp, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/terms", UseHTTPPage("terms", &fiber.Map{"Title": "Terms", "CenterContent": true}, func(r controller_http.ControllerHttp, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/privacy", UseHTTPPage("privacy", &fiber.Map{"Title": "Privacy", "CenterContent": true}, func(r controller_http.ControllerHttp, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/acknowledgements", UseHTTPPage("acknowledgements", &fiber.Map{"Title": "Acknowledgements"}, func(r controller_http.ControllerHttp, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/docs", UseHTTPPage("docs", &fiber.Map{
		"Title":          "Docs",
		"IsDocsPage":     true,
		"Docs":           docs,
		"GraphqlFields":  graphqlFields,
		"GraphqlTypes":   graphqlTypes,
		"GraphqlRequest": fieldsOf(GraphqlInput{}),
	}, func(r controller_http.ControllerHttp, bind *fiber.Map) string { return "" }, "partials/main"))
	app.Get("/settings", UseHTTPPage("settings", &fiber.Map{"Title": "Settings"},
		func(r controller_http.ControllerHttp, bind *fiber.Map) string {
			if user := r.User(); user == nil {
				return "/"
			}
			return ""
		}, "partials/main"),
	)
	app.Get("/chat", UseHTTPPage("chat", &fiber.Map{"Title": "Home", "IsChatPage": true},
		func(r controller_http.ControllerHttp, bind *fiber.Map) string {
			return ""
		}),
	)
	app.Get("/chat/groups/:group_id", UseHTTPPage("chat", &fiber.Map{"Title": "Group", "IsChatPage": true},
		func(r controller_http.ControllerHttp, bind *fiber.Map) string {
			member, _, group := r.Member()

			if member == nil {
				return "/chat"
			}

			(*bind)["Title"] = group.Nick
			return ""
		}),
	)
	app.Get("/chat/groups/join/:group_name", UseHTTPPage("chat", &fiber.Map{"Title": "Join group", "IsChatPage": true},
		func(r controller_http.ControllerHttp, bind *fiber.Map) string {
			member, group, _ := r.Member()
			if group == nil {
				return "/chat"
			}

			if member != nil {
				return controller.PathRedirectGroup(group.Id)
			}

			(*bind)["Title"] = "Join " + group.Nick
			return ""
		}),
	)

	Listen := func(method, path string, handler fiber.Handler) fiber.Router {
		return app.Add([]string{method}, path, handler)
	}

	ListenDoc := func(method, description string, fields []reflect.StructField, path string, handler fiber.Handler) fiber.Router {
		docs.HTTP[method] = append(docs.HTTP[method], DocsHTTPMethod{
			Path:        path,
			Method:      method,
			Description: description,
			Request:     fields,
			Response:    "Currently, it's always an html string response.",
		})
		return Listen(method, path, handler)
	}

	// graphql
	app.Post("/gql", func(ctx fiber.Ctx) error {
		var input GraphqlInput
		if err := ctx.Bind().Body(&input); err != nil {
			return ctx.
				Status(fiber.StatusInternalServerError).
				SendString(err.Error())
		}

		result := graphql.Do(graphql.Params{
			Schema:         schema,
			RequestString:  input.Query,
			OperationName:  input.OperationName,
			VariableValues: input.Variables,
		})

		ctx.Set("Content-Type", "application/graphql-response+json")
		return ctx.JSON(result)
	})

	// get
	ListenDoc("get", "Get messages section.", fieldsOf(model_http.MessagesPage{}), "/groups/:group_id/messages/page/:messages_page",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.MessagesPage() }))
	ListenDoc("get", "Get members section.", fieldsOf(model_http.MembersPage{}), "/groups/:group_id/members/page/:members_page",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.MembersPage() }))

	// post
	ListenDoc("post", "Create new account.", fieldsOf(model_http.UserSignUp{}), "/account/create",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserSignUp() }))
	ListenDoc("post", "Get new authorization token.", fieldsOf(model_http.UserLogin{}), "/account/login",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserLogin() }))
	ListenDoc("post", "Create new group.", fieldsOf(model_http.GroupCreate{}), "/groups/create",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.GroupCreate() }))
	ListenDoc("post", "Create (send) new message.", fieldsOf(model_http.MessageCreate{}), "/groups/:group_id/messages/create",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.MessageCreate() }))

	// put
	ListenDoc("put", "Change name identificator.", fieldsOf(model_http.UserChangeName{}), "/account/change/name",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserChangeName() }))
	ListenDoc("put", "Cahnge email.", fieldsOf(model_http.UserChangeEmail{}), "/account/change/email",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserChangeEmail() }))
	ListenDoc("put", "Change phone.", fieldsOf(model_http.UserChangePhone{}), "/account/change/phone",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserChangePhone() }))
	ListenDoc("put", "Change password.", fieldsOf(model_http.UserChangePassword{}), "/account/change/password",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserChangePassword() }))
	Listen("put", "/account/logout",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserLogout() }))
	ListenDoc("put", "Join the group immediately.", fieldsOf(model_http.GroupJoin{}), "/groups/:group_id/join",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.GroupJoin() }))
	ListenDoc("put", "Change group information and settings.", fieldsOf(model_http.GroupChange{}), "/groups/:group_id/change",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.GroupChange() }))

	// delete
	ListenDoc("delete", "Leave the group immediately.", fieldsOf(model_http.GroupLeave{}), "/groups/:group_id/leave",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.GroupLeave() }))
	ListenDoc("delete", "Delete group.", fieldsOf(model_http.GroupDelete{}), "/groups/:group_id",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.GroupDelete() }))
	ListenDoc("delete", "Delete account.", fieldsOf(model_http.UserDelete{}), "/account/delete",
		UseHTTP(func(r controller_http.ControllerHttp) error { return r.UserDelete() }))

	// websoket
	app.Get("/groups/:group_id", UseWebsocket([]string{
		controller_ws.SubForMessages,
	}, func(ws *controller_ws.ControllerWs) error { return ws.SubscribeGroup() }))

	app.Use(UseHTTPPage("partials/x", &fiber.Map{
		"Title":         fmt.Sprintf("%d", fiber.StatusNotFound),
		"StatusCode":    fiber.StatusNotFound,
		"StatusMessage": fiber.ErrNotFound.Message,
		"CenterContent": true,
	}, func(r controller_http.ControllerHttp, bind *fiber.Map) string { return "" }, "partials/main"))

	return app, nil
}
