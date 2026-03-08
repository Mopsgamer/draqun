//go:build ignore

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func git(args ...string) string {
	out, _ := exec.Command("git", args...).Output()
	return strings.TrimSpace(string(out))
}

func main() {
	data := map[string]string{
		"hash":     git("rev-parse", "--short", "HEAD"),
		"hashLong": git("rev-parse", "HEAD"),
		"branch":   git("rev-parse", "--abbrev-ref", "HEAD"),
	}

	content, _ := json.Marshal(data)
	os.MkdirAll("dist", 0755)
	os.WriteFile(filepath.Join("dist", "git.json"), content, 0644)
}
