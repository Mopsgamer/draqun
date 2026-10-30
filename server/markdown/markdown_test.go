package markdown

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdownContexts(t *testing.T) {
	input := "Hello **world**!"

	htmlMsg := string(Render(input, ContextMessage))
	assert.Contains(t, htmlMsg, "markdown-message")
	assert.Contains(t, htmlMsg, "<strong>world</strong>")

	htmlCh := string(Render(input, ContextChannelDescription))
	assert.Contains(t, htmlCh, "markdown-channel-desc")

	htmlGroup := string(Render(input, ContextGroupDescription))
	assert.Contains(t, htmlGroup, "markdown-group-desc")
}

func TestRenderCodeBlock(t *testing.T) {
	input := "```go\nfunc Hello() string\n```"
	html := string(Render(input, ContextMessage))
	t.Logf("Generated HTML:\n%s", html)
}

func TestRenderCallouts(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedClass string
		expectedTitle string
		expectedBody  string
	}{
		{
			name:          "Note Callout",
			input:         "> [!NOTE]\n> Useful info here",
			expectedClass: "callout-note",
			expectedTitle: "Note",
			expectedBody:  "Useful info here",
		},
		{
			name:          "Warning Callout",
			input:         "> [!WARNING]\n> Be careful!",
			expectedClass: "callout-warning",
			expectedTitle: "Warning",
			expectedBody:  "Be careful!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := string(Render(tt.input, ContextMessage))
			assert.Contains(t, html, "callout")
			assert.Contains(t, html, tt.expectedClass)
			assert.Contains(t, html, tt.expectedTitle)
			assert.Contains(t, html, tt.expectedBody)
			assert.NotContains(t, html, "[!NOTE]")
			assert.NotContains(t, html, "[!WARNING]")
		})
	}
}

func TestRenderHeadersListsLinksSubtext(t *testing.T) {
	// Headers
	hInput := "# Header 1\n## Header 2\n### Header 3"
	hHTML := string(Render(hInput, ContextMessage))
	assert.Contains(t, hHTML, "Header 1</h1>")
	assert.Contains(t, hHTML, "Header 2</h2>")
	assert.Contains(t, hHTML, "Header 3</h3>")

	// Lists
	lInput := "- Dash item\n* Asterisk item\n1. First item\n2. Second item"
	lHTML := string(Render(lInput, ContextMessage))
	assert.Contains(t, lHTML, "<ul>")
	assert.Contains(t, lHTML, "<li>Dash item</li>")
	assert.Contains(t, lHTML, "<li>Asterisk item</li>")
	assert.Contains(t, lHTML, "<ol>")
	assert.Contains(t, lHTML, "<li>First item</li>")

	// Links
	linkInput := "Check out [Draqun](https://draqun.example.com) and https://autolink.example.com"
	linkHTML := string(Render(linkInput, ContextMessage))
	assert.Contains(t, linkHTML, `<a href="https://draqun.example.com"`)
	assert.Contains(t, linkHTML, `>Draqun</a>`)
	assert.Contains(t, linkHTML, `<a href="https://autolink.example.com"`)

	// Discord Subtext
	subInput := "-# This is small subtext"
	subHTML := string(Render(subInput, ContextMessage))
	assert.Contains(t, subHTML, `<small class="markdown-subtext">This is small subtext</small>`)
}

func TestMaskedLinks(t *testing.T) {
	input := "Please visit [Google](https://google.com) or [Draqun Docs](https://draqun.org/docs) for details."
	html := string(Render(input, ContextMessage))

	assert.Contains(t, html, `<a href="https://google.com">Google</a>`)
	assert.Contains(t, html, `<a href="https://draqun.org/docs">Draqun Docs</a>`)
}

func TestXSSProtection(t *testing.T) {
	input := "<script>alert('xss')</script><img src=x onerror=alert(1)>"
	html := string(Render(input, ContextMessage))

	assert.NotContains(t, html, "<script>")
	assert.NotContains(t, html, "onerror=alert(1)")
}

func TestMentions(t *testing.T) {
	input := "Hello @u:1 and @u:2, welcome!\n\n```go\n// @u:3 should not be replaced\n```\n\nLink: [User @u:4](https://example.com/@u:4)"
	html := string(Render(input, ContextMessage))

	// Non-existent IDs in isolated test environment should render fallbacks safely
	assert.Contains(t, html, `@user:1`)
	assert.Contains(t, html, `@user:2`)

	// Mentions inside code blocks should NOT be altered
	assert.Contains(t, html, `@u:3`)

	// Mentions inside tag attributes should NOT be altered
	assert.Contains(t, html, `href="https://example.com/@u:4"`)
}
