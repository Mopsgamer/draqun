package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"restapp/restapp"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/template/html/v2"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := InitDB()

	app := fiber.New(fiber.Config{
		Views:             InitVE(),
		PassLocalsToViews: true,
	})

	app.Use(logger.New())

	// next code groups should be separated into different functions.
	// + should avoid code repeating

	// static
	// idk how to do it automatically, fiber
	// '/*' static methods (file hosting) wont work with js and css but with html
	app.Get("/static/js/htmx.min.js", func(c fiber.Ctx) error {
		return c.SendFile("./web/static/js/htmx.min.js")
	})
	app.Get("/static/css/main.css", func(c fiber.Ctx) error {
		return c.SendFile("./web/static/css/main.css")
	})

	// get
	app.Get("/", func(c fiber.Ctx) error {
		return c.Render("index", fiber.Map{})
	})
	app.Get("/login", func(c fiber.Ctx) error {
		return c.Render("login", fiber.Map{})
	})
	app.Get("/register", func(c fiber.Ctx) error {
		return c.Render("register", fiber.Map{})
	})

	// post
	app.Post("/register", func(c fiber.Ctx) error {
		return restapp.UserRegister(db, c)
	})
	app.Post("/login", func(c fiber.Ctx) error {
		return restapp.UserLogin(db, c)
	})

	log.Fatal(app.Listen(":3000"))
}

// also i dont like InitDB, InitVE functions inside main, so
// they can be moved into init-db.go and init-ve.go

func InitVE() *html.Engine {
	engine := html.New("./web/templates", ".html")

	engine.AddFunc("html", func(s string) template.HTML {
		return template.HTML(s)
	})

	return engine
}

func InitDB() *sqlx.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sqlx.Connect("mysql", connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to ping database: %v\n", err)
	}

	log.Println("Database connected successfully")
	return db
}
