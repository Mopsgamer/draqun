package main

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if app, err := InitServer(); err == nil {
		log.Fatal(app.Listen(":3000"))
	}
}
