package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
)

// Creates the .env file, if provided the '--init' option.
func InitProjectFiles() {

	optionInit := "--init"
	optionForce := "--force"

	path := ".env"
	isInit := slices.Contains(os.Args, optionInit)
	if !isInit {
		return
	}

	force := slices.Contains(os.Args, optionForce)
	_, errstat := os.Stat(path)
	exists := !os.IsNotExist(errstat)
	if exists && !force {
		log.Println("Failed to write " + path + " - already exists, use " + optionForce)
		os.Exit(1)
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
		os.Exit(1)
	}

	log.Println("Writed " + path)

	log.Println("Executing deno build task...")
	ExecDeno("task", "build")
	log.Println("Deno tasks completed successfully.")
	os.Exit(0)
}

// Creates the instance. Checks if the deno command exists: exits with 1 if it doesn't.
// Uses default system's output.
func ExecDeno(arg ...string) *exec.Cmd {
	_, err := exec.LookPath("deno")
	if err != nil {
		fmt.Println("Deno is not installed or not available in PATH.")
		os.Exit(1)
	}

	cmd := exec.Command("deno", arg...)

	return cmd
}

func WaitForBundleWatch() {
	deno := ExecDeno("task", "build", "--watch")

	reader, err := deno.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(reader)
	var buffer bytes.Buffer
	deno.Start()

	for scanner.Scan() {
		line := scanner.Text()

		buffer.WriteString(line + "\n")

		if strings.Contains(line, "watching") {
			break
		}
	}
}
