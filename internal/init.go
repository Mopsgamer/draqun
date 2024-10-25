package internal

import (
	"fmt"
	"html/template"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func InitServer() (*fiber.App, error) {
	db, err := InitDB()
	if err != nil {
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
	app.Get("/partials/*", func(c fiber.Ctx) error {
		r := Responder{c}
		return r.RenderTemplate()
	})

	// get
	app.Get("/", func(c fiber.Ctx) error {
		r := Responder{c}
		return r.Render("index", fiber.Map{}, "layouts/main")
	})
	app.Get("/api", func(c fiber.Ctx) error {
		r := Responder{c}
		return r.Render("api", fiber.Map{}, "layouts/main")
	})

	// post
	app.Post("/register", func(c fiber.Ctx) error {
		r := Responder{c}
		return r.UserRegister(db)
	})
	app.Post("/login", func(c fiber.Ctx) error {
		r := Responder{c}
		return r.UserLogin(db)
	})

	return app, nil
}

func InitDB() (*Database, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		return nil, err
	}

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	connection, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		return nil, err
	}

	if err := connection.Ping(); err != nil {
		log.Printf("Unable to ping database: %v\n", err)
		return nil, err
	}

	log.Println("Database connected successfully")
	return &Database{connection}, nil
}

func InitVE() *html.Engine {
	engine := html.New("./web/templates", ".html")

	engine.Reload(true)

	engine.AddFunc("html", func(s string) template.HTML {
		return template.HTML(s)
	})

	return engine
}
