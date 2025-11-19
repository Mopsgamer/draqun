package internal

import (
	"fmt"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/fatih/color"
)

func hashString() string {
	return color.New(color.Faint).Sprint(link(environment.GitHubCommit, environment.GitJson.Hash))
}

func branchString() string {
	return color.RGB(100, 0, 180).Sprint(link(environment.GitHubBranch, environment.GitJson.Branch))
}

func buildEnvironementString() string {
	return color.HiRedString(environment.BuildEnvironmentName)
}

func clientEmbeddedString(clientEmbedded bool) string {
	clientEmbeddedColor := color.RGB(0, 180, 100)
	clientEmbeddedStatus := "Client not embedded"
	if clientEmbedded {
		clientEmbeddedColor = color.New(color.FgHiRed)
		clientEmbeddedStatus = "Client embedded"
	}
	return clientEmbeddedColor.Sprint(clientEmbeddedStatus)
}

func link(link, text string) string {
	if color.NoColor {
		return text
	}
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", link, text)
}

func gitString() string {
	return branchString() + " " + hashString()
}

func versionString(clientEmbedded bool) string {
	return "v" + environment.DenoJson.Version + " " + buildEnvironementString() + " " + clientEmbeddedString(clientEmbedded)
}

func metaString(clientEmbedded bool) string {
	prefix := "│  "
	title := environment.AppName + " ─ "
	return "╭── " + color.New(color.Italic).Sprint(title) + "\n" +
		prefix + versionString(clientEmbedded) + "\n" +
		prefix + gitString() + "\n" +
		"╰──•\n"
}
