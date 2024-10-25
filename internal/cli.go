package internal

import (
	"io/fs"
	"log"
	"os"
	"slices"
)

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
