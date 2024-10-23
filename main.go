package main

import (
	"fmt"
	"log"
	"os"
	"restapp/internal/handlers"
	repository "restapp/internal/repositories"
	"restapp/internal/services"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db := InitDB()

	userRepo := repository.NewDBUserRepository(db)
	userService := services.NewUserService(userRepo)

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/", handlers.Index)
	app.Get("/login", handlers.LoginPage)
	app.Get("/register", handlers.RegisterPage)

	// maybe better rename as "auth_handler, service etc"?
	app.Post("/register", func(c fiber.Ctx) error {
		return handlers.Register(c, userService)
	})
	app.Post("/login", func(c fiber.Ctx) error {
		return handlers.Login(c, userService)
	})

	log.Fatal(app.Listen(":8080"))
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
