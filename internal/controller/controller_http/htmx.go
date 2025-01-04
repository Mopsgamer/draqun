package controller_http

import (
	"regexp"
	"strings"
)

// Otherwise json, graphql or something.
func (r *ControllerHttp) IsHTMX() bool {
	return r.Ctx.Get("HX-Request") == "true"
}

// Call it instead of Redirect().To().
func (r *ControllerHttp) HTMXRedirect(to string) {
	r.Ctx.Set("HX-Redirect", to)
}

// Refresh the page.
func (r *ControllerHttp) HTMXRefresh() {
	r.Ctx.Set("HX-Refresh", "true")
}

// Get /path/to#element?key=val
func (r *ControllerHttp) HTMXCurrentURL() string {
	return r.Ctx.Get("HX-Current-URL")
}

// Get #element
// from /path/to#element?key=val
func (r *ControllerHttp) HTMXCurrentURLHash() string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(r.HTMXCurrentURL())
}

// Get /path/to?key=val
// from /path/to#element?key=val
func (r *ControllerHttp) HTMXCurrentPath() string {
	return strings.Replace(r.HTMXCurrentURL(), r.HTMXCurrentURLHash(), "", -1)
}
