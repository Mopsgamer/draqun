package htmx

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestFormatEarlyHint(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/static/js/app.js", "</static/js/app.js>; rel=preload; as=script"},
		{"/static/css/main.css", "</static/css/main.css>; rel=preload; as=style"},
		{"/static/assets/logo.svg", "</static/assets/logo.svg>; rel=preload; as=image"},
		{"/static/assets/pic.png", "</static/assets/pic.png>; rel=preload; as=image"},
		{"/static/fonts/font.woff2", "</static/fonts/font.woff2>; rel=preload; as=font; crossorigin"},
		{"/static/unknown.xyz", "</static/unknown.xyz>; rel=preload; as=fetch"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, FormatEarlyHint(tc.path))
	}
}

func TestDefaultHintBuilder(t *testing.T) {
	app := fiber.New()

	app.Get("/test-chat", func(ctx fiber.Ctx) error {
		bind := fiber.Map{}
		hints := DefaultHintBuilder(ctx, "chat", bind)
		assert.Contains(t, hints, FormatEarlyHint(LogoPath))
		assert.Contains(t, hints, FormatEarlyHint("/static/js/app.js"))
		assert.Contains(t, hints, FormatEarlyHint("/static/css/main.css"))
		return ctx.SendStatus(fiber.StatusOK)
	})

	app.Get("/test-fallback", func(ctx fiber.Ctx) error {
		bind := fiber.Map{}
		hints := DefaultHintBuilder(ctx, "unknown-template-page", bind)
		assert.Contains(t, hints, FormatEarlyHint(LogoPath))
		assert.Contains(t, hints, FormatEarlyHint("/static/js/main.js"))
		assert.Contains(t, hints, FormatEarlyHint("/static/css/main.css"))
		return ctx.SendStatus(fiber.StatusOK)
	})

	for _, path := range []string{"/test-chat", "/test-fallback"} {
		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test failed: %v", err)
		}
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	}
}
