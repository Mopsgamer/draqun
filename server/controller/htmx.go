package controller

import (
	"regexp"
	"strings"
)

func (r *Controller) IsHTMX() bool {
	return r.Ctx.Get("HX-Request") == "true"
}

// Call it instead of Redirect().To().
func (r *Controller) HTMXRedirect(to string) {
	r.Ctx.Set("HX-Redirect", to)
}

// Refresh the page.
func (r *Controller) HTMXRefresh() {
	r.Ctx.Set("HX-Refresh", "true")
}

// Get /path/to#element?key=val
func (r *Controller) HTMXCurrentURL() string {
	return r.Ctx.Get("HX-Current-URL")
}

// Get #element
// from /path/to#element?key=val
func (r *Controller) HTMXCurrentURLHash() string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(r.HTMXCurrentURL())
}

// Get /path/to?key=val
// from /path/to#element?key=val
func (r *Controller) HTMXCurrentPath() string {
	return strings.Replace(r.HTMXCurrentURL(), r.HTMXCurrentURLHash(), "", -1)
}
