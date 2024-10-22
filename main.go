package main

import (
	"log"
	"restapp/internal/config"
	"restapp/internal/handlers"
	repository "restapp/internal/repositories"
	"restapp/internal/services"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config.InitDB()

	userRepo := repository.NewDBUserRepository(config.DBConnection)
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
