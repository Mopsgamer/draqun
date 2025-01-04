package controller_ws

import (
	"regexp"
	"restapp/internal/controller/model_ws"
	"strings"
)

// Otherwise json, graphql or something.
func (ws *ControllerWs) IsHTMX() bool {
	req := new(model_ws.WebsocketHTMX)
	err := ws.GetMessageJSON(req)
	if err != nil {
		return false
	}

	return req.HEADERS["HX-Request"] == "true"
}

// Get /path/to#element?key=val
func (ws *ControllerWs) HTMXCurrentURL() string {
	req := new(model_ws.WebsocketHTMX)
	err := ws.GetMessageJSON(req)
	if err != nil {
		return ""
	}

	return req.HEADERS["HX-Current-URL"]
}

// Get #element
// from /path/to#element?key=val
func (ws *ControllerWs) HTMXCurrentURLHash() string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(ws.HTMXCurrentURL())
}

// Get /path/to?key=val
// from /path/to#element?key=val
func (ws *ControllerWs) HTMXCurrentPath() string {
	return strings.Replace(ws.HTMXCurrentURL(), ws.HTMXCurrentURLHash(), "", -1)
}
