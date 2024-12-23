package environment

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
)

var Environment int

const (
	EnvironmentTest = iota
	EnvironmentDevelopment
	EnvironmentProduction
)

var (
	JWTKey string
	Port   string

	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
)

// Exits if any errors.
func Load() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	var err error
	environmentString := "ENVIRONMENT"
	Environment, err = strconv.Atoi(os.Getenv(environmentString))
	if err != nil {
		log.Fatalf(environmentString+" can not be '%v'. Should be an integer.", os.Getenv(environmentString))
	}
	if Environment < EnvironmentTest || Environment > EnvironmentProduction {
		log.Fatalf(environmentString+" can not be %v. Should be in the range: %v - %v.", Environment, EnvironmentTest, EnvironmentProduction)
	}
	JWTKey = os.Getenv("JWT_KEY")
	Port = os.Getenv("PORT")
	DBUser = os.Getenv("DB_USER")
	DBPassword = os.Getenv("DB_PASSWORD")
	DBName = os.Getenv("DB_NAME")
	DBHost = os.Getenv("DB_HOST")
	DBPort = os.Getenv("DB_PORT")
}

// Creates the instance. Checks if the deno command exists: exits with 1 if it doesn't.
// Uses default system's output. The instance should be started.
func ExecDeno(arg ...string) *exec.Cmd {
	_, err := exec.LookPath("deno")
	if err != nil {
		fmt.Println("Deno is not installed or not available in PATH.")
		os.Exit(1)
	}

	cmd := exec.Command("deno", arg...)

	return cmd
}

// Server can be started after this method called.
// Implements --watch and --build flags.
func WaitForBuild() {

	optionWatch := "--watch"
	optionBuild := "--build"

	isProd := Environment == EnvironmentProduction
	isBuild := slices.Contains(os.Args, optionBuild)
	isWatch := slices.Contains(os.Args, optionWatch)

	if !isBuild && !isWatch {
		if !isProd {
			log.Info("You can use --build or --watch option to bundle js, css and assets before running server.")
		}
		return
	}

	if isProd {
		log.Warn("You can use --build and --watch options only within dev environment.")
		return
	}

	if isBuild && isWatch {
		log.Fatal("Use --build or --watch, if you want to bundle while running the server. You have used both.")
		return
	}

	var deno *exec.Cmd
	if isWatch {
		log.Info("Creating file listeners for bundling js, css and assets...")
		deno = ExecDeno("task", "build", "--watch")
	} else {
		log.Info("Bundling js, css and assets...")
		deno = ExecDeno("task", "build")
	}

	reader, err := deno.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	readerErr, err := deno.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(reader)
	scannerErr := bufio.NewScanner(readerErr)
	err = deno.Start()
	if err != nil {
		log.Fatal(err)
	}

	noExit := true
	go func() {
		err := deno.Wait()
		if noExit {
			return
		}
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()
	go func() {
		for scannerErr.Scan() {
			line := scannerErr.Text()

			if strings.Contains(line, "error") || strings.Contains(line, "Error") || strings.Contains(line, "ERR") {
				noExit = false
			}
			log.Debug(line)
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()

		log.Debug(line)

		// see ./web/build.ts file
		if strings.Contains(line, "Done:") {
			log.Info("Starting the server... ENVIRONMENT = ", Environment)
			break
		}
	}

}
