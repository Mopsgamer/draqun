package internal
import (
	_ "embed"
	"fmt"
	"io/fs"
	"reflect"
	"time"
	"errors"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/controller/controller_http"
	"github.com/Mopsgamer/draqun/server/controller/controller_ws"
	"github.com/Mopsgamer/draqun/server/controller/database"
	"github.com/Mopsgamer/draqun/server/controller/model_http"
	"github.com/Mopsgamer/draqun/server/controller/model_ws"
	"github.com/Mopsgamer/draqun/server/docsgen"
	"github.com/Mopsgamer/draqun/websocket"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func CheckUser(ctx fiber.Handler) error {
	// ...
	return errors.New("Not implemented")
}

func Chain(handlers ...fiber.Handler) fiber.Handler {
	return func (ctx fiber.Ctx) error {
		for _, handler := range handlers {
			if err := handler(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// Initialize gofiber application, including DB and view engine.
func NewApp(embedFS fs.FS) (*fiber.App, error) {
	db, errDBLoad := database.InitDB()
	if errDBLoad != nil {
		log.Error(errDBLoad)
		return nil, errDBLoad
	}

	engine := NewAppHtmlEngine(db, embedFS, "client/templates")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	UseHttp := func(handler func(ctl controller_http.ControllerHttp) error) fiber.Handler {
		return func(ctx fiber.Ctx) error {
			ctl := controller_http.ControllerHttp{
				Ctx: ctx,
				DB:  *db,
			}

			x := new(model_http.MemberOfUriGroup)
			ctl.BindAll(x)
			rights, member, group, user := x.Rights(ctl)
			ctl.User = user
			ctl.Group = group
			ctl.Member = member
			ctl.Rights = rights

			return handler(ctl)
		}
	}

	UseHttpResp := func(resp controller_http.Response) fiber.Handler {
		return UseHttp(func(ctl controller_http.ControllerHttp) error {
			return resp.HandleHtmx(ctl)
		})
	}

	UseHttpPage := func(
		templatePath string,
		bind *fiber.Map,
		redirect controller_http.RedirectCompute,
		layouts ...string,
	) fiber.Handler {
		bindx := fiber.Map{
			"Title": "?",
		}
		bindx = controller.MapMerge(&bindx, bind)
		return UseHttp(func(ctl controller_http.ControllerHttp) error {
			return ctl.RenderPage(
				templatePath,
				&bindx,
				redirect,
				layouts...,
			)
		})
	}

	logWS := func(start time.Time, err error, ws *controller_ws.ControllerWs) {
		colorErr := fiber.DefaultColors.Green
		if err != nil {
			colorErr = fiber.DefaultColors.Red
		}

		fmt.Printf(
			"%s | %s%3s%s | %13s | %15s | %d | %s%s%s \n",
			time.Now().Format("15:04:05"),
			colorErr,
			"ws",
			fiber.DefaultColors.Reset,
			time.Since(start),
			ws.IP,
			ws.MessageType,
			fiber.DefaultColors.Yellow,
			ws.Message,
			fiber.DefaultColors.Reset,
		)
	}

	UseWs := func(
		subscribe []controller_ws.Subscription,
		handler func(ctl *controller_ws.ControllerWs) error,
	) fiber.Handler {
		return UseHttp(func(ctlHttp controller_http.ControllerHttp) error {
			if !websocket.IsWebSocketUpgrade(ctlHttp.Ctx) {
				return ctlHttp.Ctx.Next()
			}

			// ctlWs MUST be created before websocket handler,
			// because some methods become unavailable after protocol upgraded.
			ctlWs := controller_ws.New(ctlHttp)

			websocket.New(func(conn *websocket.Conn) {
				ctlWs.Conn = conn

				controller_ws.UserSessionMap.Connect(ctlWs.User.Id, ctlWs)
				defer controller_ws.UserSessionMap.Close(ctlWs.User.Id, ctlWs)
				for !ctlWs.Closed {
					messageType, message, err := ctlWs.Conn.ReadMessage()
					if err != nil {
						break
					}

					start := time.Now()
					ctlWs.MessageType = messageType
					ctlWs.Message = message
					err = handler(ctlWs)

					logWS(start, err, ctlWs)

					if err != nil {
						break
					}
				}
				ctlWs.Closed = true
				ctlWs.Conn.Close()
			})(ctlHttp.Ctx)

			return nil
		})
	}

	UseWsResp := func(
		subscribe []controller_ws.Subscription,
		resp controller_ws.Response,
	) fiber.Handler {
		return UseWs(subscribe, func(ctl *controller_ws.ControllerWs) error {
			return resp.HandleHtmx(ctl)
		})
	}

	UseStatic := func(dir string) fiber.Handler {
		if embedFS == nil {
			return static.New(dir, static.Config{Browse: true})
		}

		Fs, err := fs.Sub(embedFS, dir)
		if err != nil {
			log.Fatal(err)
		}

		return static.New("", static.Config{Browse: true, FS: Fs})
	}

	// static
	app.Get("/static*", UseStatic("client/static"))
	app.Get("/partials*", UseStatic("client/templates/partials"))

	// pages
	docs := docsgen.New()
	var noRedirect controller_http.RedirectCompute = func(ctl controller_http.ControllerHttp, bind *fiber.Map) string { return "" }
	var guestNoAccessRedirect controller_http.RedirectCompute = func(ctl controller_http.ControllerHttp, bind *fiber.Map) string {
		request := new(model_http.CookieUserToken)
		ctl.BindAll(request)
		user, _ := request.User(ctl)
		if user == nil {
			return "/"
		}
		return ""
	}
	app.Get("/", UseHttpPage("homepage", &fiber.Map{"Title": "Home", "IsHomePage": true}, noRedirect, "partials/main"))
	app.Get("/terms", UseHttpPage("terms", &fiber.Map{"Title": "Terms", "CenterContent": true}, noRedirect, "partials/main"))
	app.Get("/privacy", UseHttpPage("privacy", &fiber.Map{"Title": "Privacy", "CenterContent": true}, noRedirect, "partials/main"))
	app.Get("/acknowledgements", UseHttpPage("acknowledgements", &fiber.Map{"Title": "Acknowledgements"}, noRedirect, "partials/main"))
	app.Get("/docs", func(ctx fiber.Ctx) error { return ctx.Redirect().To("/docs/rest") })
	app.Get("/docs/rest", UseHttpPage("docs-rest", &fiber.Map{
		"Title":      "Rest Docs",
		"IsDocsPage": true,
		"Docs":       docs,
	}, noRedirect, "partials/main"))
	app.Get("/settings", UseHttpPage("settings", &fiber.Map{"Title": "Settings"}, guestNoAccessRedirect, "partials/main"))
	app.Get("/chat", UseHttpPage("chat", &fiber.Map{"Title": "Home", "IsChatPage": true}, noRedirect))
	app.Get("/chat/groups/:group_id", UseHttpPage("chat", &fiber.Map{"Title": "Group", "IsChatPage": true},
		func(ctl controller_http.ControllerHttp, bind *fiber.Map) string {
			request := new(model_http.MemberOfUriGroup)
			ctl.BindAll(request)
			member, group, _ := request.Member(ctl)

			if member == nil {
				return "/chat"
			}

			(*bind)["Title"] = group.Nick
			return ""
		}),
	)
	app.Get("/chat/groups/join/:group_name", UseHttpPage("chat", &fiber.Map{"Title": "Join group", "IsChatPage": true},
		func(ctl controller_http.ControllerHttp, bind *fiber.Map) string {
			request := new(model_http.MemberOfUriGroup)
			ctl.BindAll(request)
			member, group, _ := request.Member(ctl)
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

	Listen := func(method, path string, response controller_http.Response) fiber.Router {
		return app.Add([]string{method}, path, UseHttpResp(response))
	}

	var ListenDoc = func(method, description string, path string, request []reflect.StructField, response controller_http.Response) fiber.Router {
		docs.HTTP[method] = append(docs.HTTP[method], docsgen.DocsHTTPMethod{
			Path:        path,
			Method:      method,
			Description: description,
			Request:     request,
			Response:    "Currently, it's always an html string response.",
		})

		return Listen(method, path, response)
	}

	// get
	ListenDoc(
		"get",
		"Get messages section.",
		"/groups/:group_id/messages/page/:messages_page",
		docsgen.FieldsOf(model_http.MessagesPage{}),
		new(model_http.MessagesPage),
	)
	ListenDoc(
		"get",
		"Get members section.",
		"/groups/:group_id/members/page/:members_page",
		docsgen.FieldsOf(model_http.MembersPage{}),
		new(model_http.MembersPage),
	)

	// post
	ListenDoc(
		"post",
		"Create new account.",
		"/account/create",
		docsgen.FieldsOf(model_http.UserSignUp{}),
		new(model_http.UserSignUp),
	)
	ListenDoc(
		"post",
		"Get new authorization token.",
		"/account/login",
		docsgen.FieldsOf(model_http.UserLogin{}),
		new(model_http.UserLogin),
	)
	ListenDoc(
		"post",
		"Create new group.",
		"/groups/create",
		docsgen.FieldsOf(model_http.GroupCreate{}),
		new(model_http.GroupCreate),
	)
	ListenDoc(
		"post",
		"Create (send) new message.",
		"/groups/:group_id/messages/create",
		docsgen.FieldsOf(model_http.MessageCreate{}),
		new(model_http.MessageCreate),
	)

	// put
	ListenDoc(
		"put",
		"Change name identificator.",
		"/account/change/name",
		docsgen.FieldsOf(model_http.UserChangeName{}),
		new(model_http.UserChangeName),
	)
	ListenDoc(
		"put",
		"Cahnge email.",
		"/account/change/email",
		docsgen.FieldsOf(model_http.UserChangeEmail{}),
		new(model_http.UserChangeEmail),
	)
	ListenDoc(
		"put",
		"Change phone.",
		"/account/change/phone",
		docsgen.FieldsOf(model_http.UserChangePhone{}),
		new(model_http.UserChangePhone),
	)
	ListenDoc(
		"put",
		"Change password.",
		"/account/change/password",
		docsgen.FieldsOf(model_http.UserChangePassword{}),
		new(model_http.UserChangePassword),
	)
	ListenDoc(
		"put",
		"Remove authorization cookie and refresh page.",
		"/account/logout",
		docsgen.FieldsOf(model_http.UserLogout{}),
		new(model_http.UserLogout),
	)
	ListenDoc(
		"put",
		"Join the group immediately.",
		"/groups/:group_id/join",
		docsgen.FieldsOf(model_http.GroupJoin{}),
		new(model_http.GroupJoin),
	)
	ListenDoc(
		"put",
		"Change group information and settings.",
		"/groups/:group_id/change",
		docsgen.FieldsOf(model_http.GroupChange{}),
		new(model_http.GroupChange),
	)

	// delete
	ListenDoc(
		"delete",
		"Leave the group immediately.",
		"/groups/:group_id/leave",
		docsgen.FieldsOf(model_http.GroupLeave{}),
		new(model_http.GroupLeave),
	)
	ListenDoc(
		"delete",
		"Delete group.",
		"/groups/:group_id",
		docsgen.FieldsOf(model_http.GroupDelete{}),
		new(model_http.GroupDelete),
	)
	ListenDoc(
		"delete",
		"Delete account.",
		"/account/delete",
		docsgen.FieldsOf(model_http.UserDelete{}),
		new(model_http.UserDelete),
	)

	// websoket
	app.Get("/groups/:group_id", UseWsResp(
		[]controller_ws.Subscription{controller_ws.SubForMessages},
		&model_ws.WebsocketGroup{},
	))

	app.Use(UseHttpPage("partials/x", &fiber.Map{
		"Title":         fmt.Sprintf("%d", fiber.StatusNotFound),
		"StatusCode":    fiber.StatusNotFound,
		"StatusMessage": fiber.ErrNotFound.Message,
		"CenterContent": true,
	}, noRedirect, "partials/main"))

	return app, nil
}
