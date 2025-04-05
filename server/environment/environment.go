package environment

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"golang.org/x/mod/modfile"
)

const AppName string = "Draqun"
const GitHubRepo string = "https://github.com/Mopsgamer/draqun"

type EnvironmentMode int

var Environment EnvironmentMode

const (
	EnvironmentTest EnvironmentMode = iota
	EnvironmentDevelopment
	EnvironmentProduction
)

var (
	JWTKey      string
	Port        string
	DenoJson    DenoConfig
	GoMod       modfile.File
	GitHash     string
	GitHashLong string

	DBUser     string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
)

type DenoConfig struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Imports map[string]string `json:"imports"`
}

// Load environemnt variables from the '.env' file. Exits if any errors.
func Load() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	var err error
	environmentString := "ENVIRONMENT"
	environmentInt, err := strconv.Atoi(os.Getenv(environmentString))
	if err != nil {
		log.Fatalf(environmentString+" can not be '%v'. Should be an integer.", os.Getenv(environmentString))
	}
	Environment = EnvironmentMode(environmentInt)
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

	denoConfig, err := os.ReadFile("deno.json")
	if err != nil {
		log.Fatal(err)
	}

	deno := new(DenoConfig)
	err = json.Unmarshal(denoConfig, deno)
	if err != nil {
		log.Fatal(err)
	}

	DenoJson = *deno

	gomodBytes, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatal(err)
	}

	gomod, err := modfile.Parse("go.mod", gomodBytes, nil)
	if err != nil {
		log.Fatal(err)
	}

	GoMod = *gomod

	GitHash, _ = commandOutput("git", "log", "-n1", `--format="%h"`)
	GitHashLong, _ = commandOutput("git", "log", "-n1", `--format="%H"`)
}

func commandOutput(name string, arg ...string) (string, error) {
	bytes, err := exec.Command(name, arg...).Output()
	if err != nil {
		return "", err
	}

	// "hash"\n -> hash
	return string(bytes)[1 : len(bytes)-2], nil
}
