package environment

import (
	"encoding/json"
	"io"
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

var NoEnv bool

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
	GitBranch   string

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

func LoadMeta(embedFS fs.FS) {
	DenoJson = getJson[DenoConfig](embedFS, "deno.json")
	GoMod = getGoMod(embedFS)
	GitHash, _ = commandOutput("git", "rev-parse", "--short", "HEAD")
	GitHashLong, _ = commandOutput("git", "rev-parse", "HEAD")
	GitBranch, _ = commandOutput("git", "rev-parse", "--abbrev-ref", "HEAD")

	if len(GitHash) < 7 {
		log.Warn("Git hash is too short, using 'unknown' instead.")
		GitHash = "unknown"
	}
	if len(GitHashLong) < 7 {
		log.Warn("Git long hash is too short, using 'unknown' instead.")
		GitHashLong = "unknown"
	}

	if len(GitBranch) == 0 {
		log.Warn("Git branch is empty, using 'unknown' instead.")
		GitBranch = "unknown"
	}
}

// LoadEnv environemnt variables from the '.env' file. Exits if any errors.
func LoadEnv(embedFS fs.FS) {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			NoEnv = true
		} else {
			log.Error(err)
		}
	}

	JWTKey = getenvString("JWT_KEY", "")
	if len(JWTKey) < 8 {
		log.Fatal("JWT_KEY must be at least 8 characters long.")
	}
	UserAuthTokenExpiration = time.Duration(getenvInt("USER_AUTH_TOKEN_EXPIRATION", 180)) * time.Minute
	ChatMessageMaxLength = int(getenvInt("CHAT_MESSAGE_MAX_LENGTH", 8000))

	Port = getenvString("PORT", "3000")

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
	return string(bytes)[:len(bytes)-1], nil
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
// 	val := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
// 	return val == "1" || val == "true" || val == "y" || val == "yes"
// }

func getJson[T any](embedFS fs.FS, file string) T {
	fsFile, err := embedFS.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := io.ReadAll(fsFile)
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

func getGoMod(embedFS fs.FS) modfile.File {
	file := "go.mod"
	fsFile, err := embedFS.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	buf, err := io.ReadAll(fsFile)
	if err != nil {
		log.Fatal(err)
	}

	gomod, err := modfile.Parse(file, buf, nil)
	if err != nil {
		log.Fatal(err)
	}

	return *gomod
}
