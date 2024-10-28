package internal

import (
	"fmt"
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
		r := Responder{c, *db}
		return r.RenderTemplate()
	})

	// get
	app.Get("/", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage("index", "Home page", "partials/main")
	})
	app.Get("/chat", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage("chat", "Chat")
	})
	app.Get("/api", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.RenderPage("api", "API Documentation", "partials/main")
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
	app.Put("/logout", func(c fiber.Ctx) error {
		r := Responder{c, *db}
		return r.UserLogout()
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
		log.Printf("Error creating users table: %v\n", err)
		return nil, err
	}

	log.Println("Users table ensured to exist")
	return &Database{Sql: connection}, nil
}

func InitVE() *html.Engine {
	engine := html.New("./web/templates", ".html")

	engine.Reload(true)

	return engine
}
