package environment

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3/log"
	"github.com/joho/godotenv"
	"golang.org/x/mod/modfile"
)

const AppName string = "Draqun"
const GitHubRepo string = "https://github.com/Mopsgamer/draqun"
const TemplateExt = ".tmpl"
const DistFolder = "dist" // Consider using same value for go:embed comments and scripts/tool/contants.ts
const StaticFolder = DistFolder + "/static"

type BuildMode int

const (
	BuildModeDevelopment BuildMode = iota
	BuildModeProduction
)

// TODO: Should be configurable using database.
// App settings.
var (
	JWTKey                  string
	UserAuthTokenExpiration time.Duration
	ChatMessageMaxLength    int

	Port string

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
func Load(embedFS fs.FS) {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			goto InitEnv
		}
		log.Error(err)
	}

InitEnv:
	JWTKey = getenvString("JWT_KEY", "")
	UserAuthTokenExpiration = time.Duration(getenvInt("USER_AUTH_TOKEN_EXPIRATION", 180)) * time.Minute
	ChatMessageMaxLength = int(getenvInt("CHAT_MESSAGE_MAX_LENGTH", 8000))

	Port = getenvString("PORT", "3000")

	DenoJson = getJson[DenoConfig]("deno.json")
	GoMod = getGoMod()
	GitHash, _ = commandOutput("git", "log", "-n1", `--format="%h"`)
	GitHashLong, _ = commandOutput("git", "log", "-n1", `--format="%H"`)

	DBHost = getenvString("DB_HOST", "localhost")
	DBName = getenvString("DB_NAME", "mysql")
	DBPassword = getenvString("DB_PASSWORD", "")
	DBPort = getenvString("DB_PORT", "3306")
	DBUser = getenvString("DB_USER", "admin")
}

func commandOutput(name string, arg ...string) (string, error) {
	bytes, err := exec.Command(name, arg...).Output()
	if err != nil {
		return "", err
	}

	// "hash"\n -> hash
	return string(bytes)[1 : len(bytes)-2], nil
}

func getenvString(key string, or string) string {
	val := os.Getenv(key)
	if val == "" {
		return or
	}
	return val
}

func getenvInt(key string, or int64) int64 {
	val := os.Getenv(key)
	if val == "" {
		return or
	}
	result, err := strconv.ParseInt(val, 0, 64)
	if err != nil {
		log.Errorf(key+" can not be '%v'. Should be an integer.", os.Getenv(key))
		return or
	}

	return result
}

// func getenvBool(key string) bool {
// 	val := strings.ToLower(os.Getenv(key))
// 	return val == "1" || val == "true" || val == "y" || val == "yes"
// }

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
