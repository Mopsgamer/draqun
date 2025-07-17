package htmx

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func IsHtmx(ctx fiber.Ctx) bool {
	return ctx.Get("HX-Request") == "true"
}

// Set HX-Redirect header.
func Redirect(ctx fiber.Ctx, to string) {
	ctx.Set("HX-Redirect", to)
}

// Enable HX-Refresh header.
func EnableRefresh(ctx fiber.Ctx) {
	ctx.Set("HX-Refresh", "true")
}

// Get HX-Current-URL header "/path/to#element?key=val".
func Url(ctx fiber.Ctx) string {
	return ctx.Get("HX-Current-URL")
}

// Get "#element" from Url "/path/to#element?key=val".
func UrlHash(ctx fiber.Ctx) string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(Url(ctx))
}

// Get "/path/to?key=val" from Url "/path/to#element?key=val".
func Path(ctx fiber.Ctx) string {
	return strings.Replace(Url(ctx), UrlHash(ctx), "", -1)
}
