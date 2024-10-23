package main

import (
	"html/template"

	"github.com/gofiber/template/html/v2"
)

func InitVE() *html.Engine {
	engine := html.New("./web/templates", ".html")

	engine.Reload(true)

	engine.AddFunc("html", func(s string) template.HTML {
		return template.HTML(s)
	})

	return engine
}
