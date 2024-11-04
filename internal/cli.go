package internal

import (
	"io/fs"
	"log"
	"os"
	"slices"
)

// Creates the .env file, if provided the '--init' option.
func InitProject() (hasOption bool) {

	optionInit := "--init"
	optionForce := "--force"

	path := ".env"
	isInit := slices.Contains(os.Args, optionInit)
	if !isInit {
		return isInit
	}

	force := slices.Contains(os.Args, optionForce)
	_, errstat := os.Stat(path)
	exists := !os.IsNotExist(errstat)
	if exists && !force {
		log.Println("Failed to write " + path + " - already exists, use " + optionForce)
		return isInit
	}

	err := os.WriteFile(path, []byte(
		"DB_PASSWORD=\n"+
			"JWT_KEY=\n"+
			"DB_NAME=restapp\n"+
			"DB_USER=root\n"+
			"DB_HOST=localhost\n"+
			"DB_PORT=3306\n",
	), fs.ModeDevice)

	if err != nil {
		log.Println("Failed to write " + path)
		return isInit
	}

	log.Println("Writed " + path)
	return isInit
}
