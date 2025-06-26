//go:build lite

package main

import (
	"embed"

	server "github.com/Mopsgamer/draqun/server"
)

//go:embed go.mod deno.json
var embedFS embed.FS

func main() {
	server.Serve(embedFS)
}
