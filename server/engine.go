package internal

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/Mopsgamer/draqun/server/controller"
	"github.com/Mopsgamer/draqun/server/database"
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/Mopsgamer/draqun/server/model_database"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v2"
)

// Initialize the view engine.
func NewAppHtmlEngine(db *database.Database, embedFS fs.FS, clientEmbedded bool, directory string) *html.Engine {
	var engine *html.Engine
	if !clientEmbedded {
		engine = html.New(directory, environment.TemplateExt)
	} else {
		embedTemplates, _ := fs.Sub(embedFS, directory)
		engine = html.NewFileSystem(http.FS(embedTemplates), environment.TemplateExt)
	}

	if environment.BuildModeValue == environment.BuildModeDevelopment {
		engine.Reload(true)
	}

	engine.AddFuncMap(map[string]any{
		"add": func(v ...int) int {
			result := 0
			for _, num := range v {
				result += num
			}
			return result
		},
		"hideEmail": func(text string) string {
			splits := strings.Split(text, "@")
			if len(splits) != 2 {
				return strings.Repeat("*", len(text))
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
			return t.Format("2006-01-02T15:04:05.000Z")
		},
		"hidePhone": func(text string) string {
			if len(text) > 5 {
				return text[:4] + strings.Repeat("*", len(text)-4)
			}
			return strings.Repeat("*", len(text))
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

		"groupLink": func(group model_database.Group) string {
			return "localhost:3000" + controller.PathRedirectGroupJoin(group.Name)
		},
		"userRightsOf":    db.MemberRights,
		"userMemberOf":    db.MemberById,
		"userMemberships": db.UserGroupList,
		"groupMembers":    db.MemberList,
	})

	return engine
}

func satisfies[T any](v any) bool {
	_, ok := v.(T)
	return ok
}
