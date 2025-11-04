//go:build !lite

package main

import (
	"embed"
)

//go:embed go.mod deno.json dist/git.json dist/static/** client/templates/**
var embedFS embed.FS

const isClientEmbedded = true
