package logic_websocket

import (
	"regexp"
	"restapp/internal/logic/model_request"
	"strings"
)

// Otherwise json, graphql or something.
func (ws *LogicWebsocket) IsHTMX() bool {
	req := new(model_request.WebsocketHTMX)
	err := ws.GetMessageJSON(req)
	if err != nil {
		return false
	}

	return req.HEADERS["HX-Request"] == "true"
}

// Get /path/to#element?key=val
func (ws *LogicWebsocket) HTMXCurrentURL() string {
	req := new(model_request.WebsocketHTMX)
	err := ws.GetMessageJSON(req)
	if err != nil {
		return ""
	}

	return req.HEADERS["HX-Current-URL"]
}

// Get #element
// from /path/to#element?key=val
func (ws *LogicWebsocket) HTMXCurrentURLHash() string {
	return regexp.MustCompile(`((#[a-zA-Z0-9_-]+)|(\?[a-zA-Z_]))+`).FindString(ws.HTMXCurrentURL())
}

// Get /path/to?key=val
// from /path/to#element?key=val
func (ws *LogicWebsocket) HTMXCurrentPath() string {
	return strings.Replace(ws.HTMXCurrentURL(), ws.HTMXCurrentURLHash(), "", -1)
}
