//go:build lite

package main

import (
	"embed"
)

//go:embed go.mod dist/git.json package.json
var embedFS embed.FS

const isClientEmbedded = false
