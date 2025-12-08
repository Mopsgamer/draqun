package internal

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v3"
)

// Initialize the view engine.
func NewAppHtmlEngine(embedFS fs.FS, clientEmbedded bool, directory string) *html.Engine {
	var engine *html.Engine
	if !clientEmbedded {
		engine = html.New(directory, environment.TemplateExt)
	} else {
		embedTemplates, _ := fs.Sub(embedFS, directory)
		engine = html.NewFileSystem(http.FS(embedTemplates), environment.TemplateExt)
	}

	if environment.BuildEnvironment == environment.BuildModeDevelopment {
		engine.Reload(true)
	}

	engine.AddFuncMap(map[string]any{
		"add": func(v ...uint) (result uint) {
			for _, num := range v {
				result += num
			}
			return result
		},
		"hideEmail": func(email model.Email) string {
			splits := strings.Split(string(email), "@")
			if len(splits) != 2 {
				return strings.Repeat("*", len(email))
			}
			// a in a@b.c
			before := splits[0]
			// @b.c in a@b.c
			after := "@" + splits[1]

			if len(before) > 5 {
				before = before[:3] + strings.Repeat("*", len(before[3:]))
			} else {
				before = strings.Repeat("*", len(before))
			}
			return before + after
		},
		"jsonTime": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
		"hidePhone": func(phone model.Phone) string {
			if len(phone) > 5 {
				return string(phone)[:4] + strings.Repeat("*", len(phone)-4)
			}
			return strings.Repeat("*", len(phone))
		},
		"hide": func(text string) string {
			return strings.Repeat("*", len(text))
		},
		"isString": satisfies[string],
		"newMap": func(args ...any) fiber.Map {
			result := fiber.Map{}
			for i := 0; i < len(args)-1; i = i + 2 {
				k := args[i].(string)
				v := args[i+1]
				result[k] = v
			}
			return result
		},
	})

	return engine
}

func satisfies[T any](v any) bool {
	_, ok := v.(T)
	return ok
}
