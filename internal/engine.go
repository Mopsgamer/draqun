package internal

import (
	"restapp/internal/environment"
	"restapp/internal/model"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/template/html/v2"
)

// Initialize the view engine.
func NewAppHtmlEngine(db *Database) *html.Engine {
	engine := html.New("./web/templates", ".html")

	if environment.Environment == environment.EnvironmentDevelopment {
		engine.Reload(true)
	}

	engine.AddFuncMap(map[string]interface{}{
		"paginateGroups": paginate[model.Group],
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
		"hidePhone": func(text string) string {
			if len(text) > 5 {
				return text[:4] + strings.Repeat("*", len(text)-4)
			}
			return strings.Repeat("*", len(text))
		},
		"hide": func(text string) string {
			return strings.Repeat("*", len(text))
		},

		// TODO: implement cache
		"memberOf":   db.UserGroupList,
		"membersOf":  db.GroupMemberList,
		"messagesOf": db.GroupMessageList,
		"userById":   db.UserById,
		"groupById":  db.GroupById,
	})

	return engine
}

func paginate[T any](slice []T, n int) [][]T {
	result := [][]T{}
	for v := range slices.Chunk(slice, n) {
		result = append(result, v)
	}

	log.Info(result)
	return result
}
