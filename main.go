//go:generate go run scripts/gen_git.go
package main

import (
	server "github.com/Mopsgamer/draqun/server"
)

func main() {
	server.Serve(embedFS, isClientEmbedded)
}
