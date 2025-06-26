//go:build !lite

package main

import (
	"embed"

	server "github.com/Mopsgamer/draqun/server"
)

//go:embed dist/static/** client/templates/**
var embedFS embed.FS

func main() {
	server.Serve(embedFS)
}
