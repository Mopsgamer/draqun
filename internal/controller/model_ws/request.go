package model_ws

import (
	"regexp"
	"restapp/internal/controller/controller_ws"
	"strings"
)

type Request struct {
	CookieUserToken
	HEADERS map[string]string `json:"HEADERS"`
}

// Check if the request was sended by htmx library.
func (req Request) IsHTMX(ctl controller_ws.ControllerWs) bool {
	err := ctl.GetMessageJSON(req)
	if err != nil {
		return false
	}

	return req.HEADERS["HX-Request"] == "true"
}

// Get /path/to#element?key=val
func (req Request) HTMXCurrentURL(ctl controller_ws.ControllerWs) string {
	err := ctl.GetMessageJSON(req)
	if err != nil {
		return ""
	}

	return req.HEADERS["HX-Current-URL"]
}

// Get #element
// from /path/to#element?key=val
func (req Request) HTMXCurrentURLHash(ctl controller_ws.ControllerWs) string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(req.HTMXCurrentURL(ctl))
}

// Get /path/to?key=val
// from /path/to#element?key=val
func (req Request) HTMXCurrentPath(ctl controller_ws.ControllerWs) string {
	return strings.Replace(req.HTMXCurrentURL(ctl), req.HTMXCurrentURLHash(ctl), "", -1)
}
