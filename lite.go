//go:build lite

package main

import (
	server "github.com/Mopsgamer/draqun/server"
)

func main() {
	server.Serve(nil)
}
