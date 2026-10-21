package environment

import (
	"crypto/rand"
	"encoding/hex"
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
const DistFolder = "dist" // Consider using same value for go:embed comments and scripts/tool/constants.ts
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
	PkgJson      PackageJson
	GitJson      GitInfo
	GoMod        modfile.File

	DBPath string
)

type PackageJson struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}

type GitInfo struct {
	Hash     string `json:"hash"`
	HashLong string `json:"hashlong"`
	Branch   string `json:"branch"`
}

func LoadMeta(embedFS fs.FS) {
	PkgJson, _ = getJson[PackageJson](embedFS, "package.json")
	GitJson, _ = getJson[GitInfo](embedFS, DistFolder+"/git.json")
	GoMod, _ = getGoMod(embedFS)

	if GitJson.HashLong != "" {
		GitHubCommit = GitHubRepo + "/commit/" + GitJson.HashLong
	}
	if GitJson.Branch != "" {
		GitHubBranch = GitHubRepo + "/tree/" + GitJson.Branch
	}
}

func generateRandomKey(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func ensureEnvFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	randomJWT := generateRandomKey(16)
	content := `# use any online jwt generator to fill this value:
# - https://randomfungenerator.com/generators/jwt-generator
# default: <empty>
JWT_KEY=` + randomJWT + `

# in minutes
# default: 180
USER_AUTH_TOKEN_EXPIRATION=180

# max characters quantity after spaces are trimed
# default: 8000
CHAT_MESSAGE_MAX_LENGTH=8000

# application port
# default: 3000
PORT=3000

# local sqlite database file path
# default: app_data.db
DB_PATH=app_data.db
`
	return os.WriteFile(path, []byte(content), 0644)
}

// LoadEnv environemnt variables from the '.env' file. Exits if any errors.
func LoadEnv(embedFS fs.FS) (errEnv error) {
	_ = ensureEnvFile(".env")

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

	DBPath = getenvString("DB_PATH", "app_data.db")
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
