package environment

import (
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"golang.org/x/mod/modfile"
)

const AppName string = "Draqun"
const GitHubRepo string = "https://github.com/Mopsgamer/draqun"

type BuildMode int

const (
	BuildModeTest BuildMode = iota
	BuildModeDevelopment
	BuildModeProduction
)

// TODO: Should be configurable using database.
// App settings.
var (
	JWTKey                  string
	UserAuthTokenExpiration time.Duration
	ChatMessageMaxLength    int

	Environment BuildMode
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

	JWTKey = os.Getenv("JWT_KEY")
	UserAuthTokenExpiration = time.Duration(getenvInt("USER_AUTH_TOKEN_EXPIRATION")) * time.Minute
	ChatMessageMaxLength = int(getenvInt("CHAT_MESSAGE_MAX_LENGTH"))

	Environment = BuildMode(getenvInt("ENVIRONMENT"))
	Port = os.Getenv("PORT")

	DenoJson = getJson[DenoConfig]("deno.json")
	GoMod = getGoMod()
	GitHash, _ = commandOutput("git", "log", "-n1", `--format="%h"`)
	GitHashLong, _ = commandOutput("git", "log", "-n1", `--format="%H"`)

	DBHost = os.Getenv("DB_HOST")
	DBName = os.Getenv("DB_NAME")
	DBPassword = os.Getenv("DB_PASSWORD")
	DBPort = os.Getenv("DB_PORT")
	DBUser = os.Getenv("DB_USER")
}

func commandOutput(name string, arg ...string) (string, error) {
	bytes, err := exec.Command(name, arg...).Output()
	if err != nil {
		return "", err
	}

	// "hash"\n -> hash
	return string(bytes)[1 : len(bytes)-2], nil
}

func getenvInt(key string) int64 {
	val := os.Getenv(key)
	result, err := strconv.ParseInt(val, 0, 64)
	if err != nil {
		log.Fatalf(key+" can not be '%v'. Should be an integer.", os.Getenv(key))
	}

	return result
}

func getenvBool(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "y" || val == "yes"
}

func getJson[T any](file string) T {
	buf, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	val := new(T)
	err = json.Unmarshal(buf, val)
	if err != nil {
		log.Fatal(err)
	}

	return *val
}

func getGoMod() modfile.File {
	buf, err := os.ReadFile("go.mod")
	if err != nil {
		log.Fatal(err)
	}

	gomod, err := modfile.Parse("go.mod", buf, nil)
	if err != nil {
		log.Fatal(err)
	}

	return *gomod
}
