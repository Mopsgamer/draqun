package main

import (
	"io/fs"
	"log"
	"os"
	"slices"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	if CreateEnv("--make-env") {
		return
	}
	if app, err := InitServer(); err == nil {
		log.Fatal(app.Listen(":3000"))
	}
}

func CreateEnv(option string) (hasOption bool) {
	path := ".env"
	makeEnv := slices.Contains(os.Args, option)
	if !makeEnv {
		return makeEnv
	}

	force := slices.Contains(os.Args, "--force")
	_, errstat := os.Stat(path)
	exists := !os.IsNotExist(errstat)
	if exists && !force {
		log.Println("Failed to write " + path + " - already exists, use --force")
		return makeEnv
	}

	err := os.WriteFile(path, []byte(
		"DB_PASSWORD=\n"+
			"DB_NAME=restapp\n"+
			"DB_USER=root\n"+
			"DB_HOST=localhost\n"+
			"DB_PORT=3306\n",
	), fs.ModeDevice)
	if err != nil {
		log.Println("Failed to write " + path)
		return makeEnv
	}
	log.Println("Writed " + path)
	return makeEnv
}
