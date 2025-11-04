//go:build lite

package main

import (
	"embed"
)

//go:embed go.mod dist/git.json deno.json
var embedFS embed.FS

const isClientEmbedded = false
