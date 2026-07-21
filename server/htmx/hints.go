package htmx

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

type HintBuilder func(ctx fiber.Ctx, name string, bind fiber.Map) []string

func FormatEarlyHint(path string) string {
	asType := "fetch"
	lowerPath := strings.ToLower(path)

	if strings.HasSuffix(lowerPath, ".js") {
		asType = "script"
	} else if strings.HasSuffix(lowerPath, ".css") {
		asType = "style"
	} else if strings.HasSuffix(lowerPath, ".svg") ||
		strings.HasSuffix(lowerPath, ".png") ||
		strings.HasSuffix(lowerPath, ".jpg") ||
		strings.HasSuffix(lowerPath, ".jpeg") ||
		strings.HasSuffix(lowerPath, ".webp") ||
		strings.HasSuffix(lowerPath, ".gif") ||
		strings.HasSuffix(lowerPath, ".ico") {
		asType = "image"
	} else if strings.HasSuffix(lowerPath, ".woff") ||
		strings.HasSuffix(lowerPath, ".woff2") ||
		strings.HasSuffix(lowerPath, ".ttf") ||
		strings.HasSuffix(lowerPath, ".otf") {
		asType = "font"
	}

	hint := "<" + path + ">; rel=preload; as=" + asType
	if asType == "font" {
		hint += "; crossorigin"
	}
	return hint
}

const LogoPath = "/static/assets/logo.svg"

var DefaultHints = []string{"/static/js/main.js", "/static/css/main.css"}

var TemplateHints = map[string][]string{
	"homepage":        {"/static/js/homepage.js", "/static/css/homepage.css"},
	"docs":            {"/static/js/docs.js", "/static/css/docs.css"},
	"chat":            {"/static/js/app.js", "/static/css/main.css"},
	"chat-login":      {"/static/js/app.js", "/static/css/main.css"},
	"chat-group":      {"/static/js/app.js", "/static/css/main.css"},
	"chat-group-join": {"/static/js/app.js", "/static/css/main.css"},
}

var CurrentHintBuilder HintBuilder = DefaultHintBuilder

func DefaultHintBuilder(ctx fiber.Ctx, name string, bind fiber.Map) []string {
	var paths []string

	if registered, ok := TemplateHints[name]; ok {
		paths = registered
	} else {
		paths = DefaultHints
	}

	rawPaths := append([]string{LogoPath}, paths...)

	hints := make([]string, len(rawPaths))
	for i, path := range rawPaths {
		hints[i] = FormatEarlyHint(path)
	}

	return hints
}
