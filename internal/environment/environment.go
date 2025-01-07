package environment

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"golang.org/x/mod/modfile"
)

var Environment int

const (
	EnvironmentTest = iota
	EnvironmentDevelopment
	EnvironmentProduction
)

var (
	JWTKey   string
	Port     string
	DenoJson DenoConfig
	GoMod    modfile.File

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
}
