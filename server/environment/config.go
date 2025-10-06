package environment

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
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

	GitHubCommit = ""
	GitHubBranch = ""
	DenoJson     DenoConfig
	GitJson      GitInfo
	GoMod        modfile.File

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

type GitInfo struct {
	Hash     string `json:"hash"`
	HashLong string `json:"hashlong"`
	Branch   string `json:"branch"`
}

func LoadMeta(embedFS fs.FS) {
	DenoJson, _ = getJson[DenoConfig](embedFS, "deno.json")
	GitJson, _ = getJson[GitInfo](embedFS, DistFolder+"/git.json")
	GoMod, _ = getGoMod(embedFS)

	GitHubCommit = GitHubRepo + "/commit/" + GitJson.HashLong
	GitHubBranch = GitHubRepo + "/tree/" + GitJson.Branch
}

// LoadEnv environemnt variables from the '.env' file. Exits if any errors.
func LoadEnv(embedFS fs.FS) (errEnv error) {
	errEnv = nil
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			NoEnv = true
		} else {
			errEnv = err
		}
	}

	JWTKey = getenvString("JWT_KEY", "")
	if len(JWTKey) < 8 {
		return errors.New("JWT_KEY must be at least 8 characters long")
	}
	UserAuthTokenExpiration = time.Duration(getenvInt("USER_AUTH_TOKEN_EXPIRATION", 180)) * time.Minute
	ChatMessageMaxLength = int(getenvInt("CHAT_MESSAGE_MAX_LENGTH", 8000))

	Port = getenvString("PORT", "3000")

	DBHost = getenvString("DB_HOST", "localhost")
	DBName = getenvString("DB_NAME", "mysql")
	DBPassword = getenvString("DB_PASSWORD", "")
	DBPort = getenvString("DB_PORT", "3306")
	DBUser = getenvString("DB_USER", "admin")
	return errEnv
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

func getJson[T any](embedFS fs.FS, file string) (val T, err error) {
	fsFile, err := embedFS.Open(file)
	if err != nil {
		return val, err
	}

	buf, err := io.ReadAll(fsFile)
	if err != nil {
		return val, err
	}

	err = json.Unmarshal(buf, &val)
	if err != nil {
		return val, err
	}

	return val, nil
}

func getGoMod(embedFS fs.FS) (val modfile.File, err error) {
	file := "go.mod"
	fsFile, err := embedFS.Open(file)
	if err != nil {
		return val, err
	}

	buf, err := io.ReadAll(fsFile)
	if err != nil {
		return val, err
	}

	gomod, err := modfile.Parse(file, buf, nil)
	if err != nil {
		return val, err
	}

	return *gomod, nil
}
