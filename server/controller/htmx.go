package controller

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func IsHTMX(ctx fiber.Ctx) bool {
	return ctx.Get("HX-Request") == "true"
}

// Call it instead of Redirect().To().
func HTMXRedirect(ctx fiber.Ctx, to string) {
	ctx.Set("HX-Redirect", to)
}

// Refresh the page.
func HTMXRefresh(ctx fiber.Ctx) {
	ctx.Set("HX-Refresh", "true")
}

// Get /path/to#element?key=val
func HTMXCurrentURL(ctx fiber.Ctx) string {
	return ctx.Get("HX-Current-URL")
}

// Get #element
// from /path/to#element?key=val
func HTMXCurrentURLHash(ctx fiber.Ctx) string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(HTMXCurrentURL(ctx))
}

// Get /path/to?key=val
// from /path/to#element?key=val
func HTMXCurrentPath(ctx fiber.Ctx) string {
	return strings.Replace(HTMXCurrentURL(ctx), HTMXCurrentURLHash(ctx), "", -1)
}
