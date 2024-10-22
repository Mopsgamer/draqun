package main

import (
	"html/template"

	"github.com/gofiber/template/html/v2"
)

// Using templates: https://github.com/gofiber/template/blob/master/html/README.md

func declFuncs(engine *html.Engine) {
	engine.AddFunc("html", func(s string) template.HTML {
		return template.HTML(s)
	})
}
