package internal

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
)

// Initialize gofiber application, including DB and view engine.
func InitServer() (*fiber.App, error) {
	WaitForBundleWatch()

	db, err := InitDB()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	app := fiber.New(fiber.Config{
		Views:             InitVE(),
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	// next code groups should be separated into different functions.
	// + should avoid code repeating

	// static
	app.Get("/static/*", static.New("./web/static"))
	app.Get("/assets/*", static.New("./web/assets"))
	app.Get("/partials/*", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderTemplate()
	})

	// get
	app.Get("/", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"",
			"index",
			fiber.Map{
				"Title": "Restapp - Home page",
			},
			"partials/main",
		)
	})
	app.Get("/settings", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"/",
			"settings",
			fiber.Map{
				"Title": "Restapp - Settings",
			},
			"partials/main",
		)
	})
	app.Get("/chat", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"",
			"chat",
			fiber.Map{
				"Title":      "Restapp - Chat",
				"IsChatPage": true,
			},
		)
	})
	app.Get("/api", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage(
			"",
			"api",
			fiber.Map{
				"Title": "API Docs",
			},
			"partials/main",
		)
	})

	// post
	app.Post("/register", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserRegister()
	})
	app.Post("/login", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogin()
	})

	// put
	app.Put("/change-name", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangeName()
	})
	app.Put("/change-email", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangeEmail()
	})
	app.Put("/change-phone", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangePhone()
	})
	app.Put("/change-password", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserChangePassword()
	})
	app.Put("/logout", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogout()
	})

	// delete
	app.Delete("/account-delete", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserDelete()
	})

	return app, nil
}

// Initialize the DB wrapper.
func InitDB() (*Database, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	connection, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	if err := connection.Ping(); err != nil {
		return nil, err
	}

	log.Info("Database connected successfully")

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INT NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
		name VARCHAR(255) DEFAULT NULL,
		tag VARCHAR(255) DEFAULT NULL,
		email VARCHAR(255) NOT NULL,
		phone VARCHAR(255) DEFAULT NULL,
		password VARCHAR(255) NOT NULL,
		avatar VARCHAR(255) DEFAULT NULL,
		created_at DATETIME DEFAULT NULL COMMENT 'Account create time',
		last_seen DATETIME DEFAULT NULL COMMENT 'Last seen time',
		registered TINYINT(1) DEFAULT '0',
		PRIMARY KEY (id)
	) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='Users data'
`

	if _, err := connection.Exec(createTableQuery); err != nil {
		return nil, err
	}

	log.Info("Users table ensured to exist")
	return &Database{Sql: connection}, nil
}

// Initialize the view engine.
func InitVE() *html.Engine {
	engine := html.New("./web/templates", ".html")

	engine.Reload(true)

	engine.AddFuncMap(map[string]interface{}{
		"hideEmail": func(text string) string {
			splits := strings.Split(text, "@")
			if len(splits) != 2 {
				return template.HTMLEscapeString(text)
			}
			// a in a@b.c
			before := splits[0]
			// @b.c in a@b.c
			after := "@" + splits[1]

			if len(before) > 5 {
				before = before[:3] + before[3:]
			} else {
				before = strings.Repeat("*", len(before))
			}
			return before + after
		},
		"hidePhone": func(text string) string {
			if len(text) > 5 {
				return text[:4] + strings.Repeat("*", len(text)-4)
			}
			return strings.Repeat("*", len(text))
		},
		"hide": func(text string) string {
			return strings.Repeat("*", len(text))
		},
	})

	return engine
}
