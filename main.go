package main

import (
	"log"
	"restapp/internal"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if internal.CreateEnv("--make-env") {
		return
	}
	if app, err := internal.InitServer(); err == nil {
		log.Fatal(app.Listen(":3000"))
	}
}
