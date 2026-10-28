package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type ContextType string

const (
	ContextMessage            ContextType = "message"
	ContextChannelDescription ContextType = "channel-description"
	ContextGroupDescription   ContextType = "group-description"
)

var (
	calloutMarkerRegex       = regexp.MustCompile(`(?m)^(\s*>\s*)\[\!(NOTE|WARNING|TIP|IMPORTANT|CAUTION)\][ \t]*`)
	calloutMatchRegex        = regexp.MustCompile(`(?i)%%CALLOUT:(NOTE|WARNING|TIP|IMPORTANT|CAUTION)%%`)
	discordSubtextRegex      = regexp.MustCompile(`(?m)^(\s*)-\#\s+(.+)$`)
	subtextPlaceholderRegex = regexp.MustCompile(`%%SUBTEXT_START%%(.*?)%%SUBTEXT_END%%`)
	mentionRegex            = regexp.MustCompile(`(?:^|[^a-zA-Z0-9_/\.-])@([a-zA-Z0-9._]{1,64})\b`)
	codeBlockWrapperRegex   = regexp.MustCompile(`(?s)<div class="code-block-wrapper">.*?</div>|<pre.*?>.*?</pre>`)
	htmlTagRegex            = regexp.MustCompile(`<[^>]+>`)
)

func preprocessMarkdown(input string) string {
	// 1. Process callouts: replace "> [!NOTE]" with "> %%CALLOUT:NOTE%%"
	processed := calloutMarkerRegex.ReplaceAllString(input, "${1}%%CALLOUT:${2}%%")
	// 2. Process Discord subtext syntax safely via placeholders so raw HTML remains disabled
	processed = discordSubtextRegex.ReplaceAllString(processed, "${1}%%SUBTEXT_START%%${2}%%SUBTEXT_END%%")
	return processed
}

func postprocessHTML(htmlInput string) string {
	htmlInput = subtextPlaceholderRegex.ReplaceAllString(htmlInput, `<small class="markdown-subtext">${1}</small>`)

	// 1. Extract code blocks so mentions inside code snippets are protected
	var codeBlocks []string
	protected := codeBlockWrapperRegex.ReplaceAllStringFunc(htmlInput, func(m string) string {
		idx := len(codeBlocks)
		codeBlocks = append(codeBlocks, m)
		return fmt.Sprintf("%%PROTECTED_CODEBLOCK_%d%%", idx)
	})

	// 2. Extract HTML tags so mentions inside tag attributes (href, src) are protected
	var htmlTags []string
	protected = htmlTagRegex.ReplaceAllStringFunc(protected, func(m string) string {
		idx := len(htmlTags)
		htmlTags = append(htmlTags, m)
		return fmt.Sprintf("%%PROTECTED_TAG_%d%%", idx)
	})

	// 3. Process mentions in plain text
	processedText := mentionRegex.ReplaceAllStringFunc(protected, func(m string) string {
		match := mentionRegex.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		prefix := m[:strings.Index(m, "@")]
		username := match[1]
		badge := fmt.Sprintf(`<button type="button" class="mention-badge" hx-get="/users/by-name/%s/profile" hx-target="#user-profile-modal-content" onclick="document.getElementById('user-profile-modal').showModal()">@%s</button>`, username, username)
		return prefix + badge
	})

	// 4. Restore HTML tags
	for i, tag := range htmlTags {
		placeholder := fmt.Sprintf("%%PROTECTED_TAG_%d%%", i)
		processedText = strings.ReplaceAll(processedText, placeholder, tag)
	}

	// 5. Restore code blocks
	for i, codeBlock := range codeBlocks {
		placeholder := fmt.Sprintf("%%PROTECTED_CODEBLOCK_%d%%", i)
		processedText = strings.ReplaceAll(processedText, placeholder, codeBlock)
	}

	return processedText
}

type calloutInfo struct {
	Kind  string // NOTE, WARNING, TIP, IMPORTANT, CAUTION
	Title string
	Class string
	Icon  string
}

func getCalloutInfo(kind string) calloutInfo {
	switch strings.ToUpper(kind) {
	case "NOTE":
		return calloutInfo{
			Kind:  "NOTE",
			Title: "Note",
			Class: "callout-note",
			Icon:  "icon-[lucide--info]",
		}
	case "WARNING":
		return calloutInfo{
			Kind:  "WARNING",
			Title: "Warning",
			Class: "callout-warning",
			Icon:  "icon-[lucide--triangle-alert]",
		}
	case "TIP":
		return calloutInfo{
			Kind:  "TIP",
			Title: "Tip",
			Class: "callout-tip",
			Icon:  "icon-[lucide--lightbulb]",
		}
	case "IMPORTANT":
		return calloutInfo{
			Kind:  "IMPORTANT",
			Title: "Important",
			Class: "callout-important",
			Icon:  "icon-[lucide--alert-circle]",
		}
	case "CAUTION":
		return calloutInfo{
			Kind:  "CAUTION",
			Title: "Caution",
			Class: "callout-caution",
			Icon:  "icon-[lucide--shield-alert]",
		}
	default:
		return calloutInfo{
			Kind:  "NOTE",
			Title: "Note",
			Class: "callout-note",
			Icon:  "icon-[lucide--info]",
		}
	}
}

type customHTMLRenderer struct {
	renderer.NodeRenderer
}

func newCustomHTMLRenderer() renderer.NodeRenderer {
	return &customHTMLRenderer{}
}

func (r *customHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
}

func (r *customHTMLRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	bq := node.(*ast.Blockquote)

	var calloutKind string
	if firstPara, ok := bq.FirstChild().(*ast.Paragraph); ok {
		if firstPara.Lines().Len() > 0 {
			firstLine := firstPara.Lines().At(0)
			lineVal := string(firstLine.Value(source))
			match := calloutMatchRegex.FindStringSubmatchIndex(lineVal)
			if len(match) > 3 {
				calloutKind = lineVal[match[2]:match[3]]
				if entering {
					matchEnd := match[1]
					for child := firstPara.FirstChild(); child != nil; {
						next := child.NextSibling()
						if txt, isTxt := child.(*ast.Text); isTxt {
							txtVal := string(txt.Segment.Value(source))
							if strings.Contains(txtVal, "%%CALLOUT:") {
								if txt.Segment.Len() <= matchEnd {
									firstPara.RemoveChild(firstPara, txt)
								} else {
									txt.Segment = text.NewSegment(txt.Segment.Start+matchEnd, txt.Segment.Stop)
								}
							}
						}
						child = next
					}
				}
			}
		}
	}

	if calloutKind != "" {
		if entering {
			info := getCalloutInfo(calloutKind)
			fmt.Fprintf(w, `<div class="callout %s"><div class="callout-header"><span class="icon %s size-4 shrink-0" aria-hidden="true"></span><span class="callout-title">%s</span></div><div class="callout-body">`,
				info.Class, info.Icon, info.Title)
		} else {
			_, _ = w.WriteString("</div></div>")
		}
		return ast.WalkContinue, nil
	}

	if entering {
		_, _ = w.WriteString("<blockquote>")
	} else {
		_, _ = w.WriteString("</blockquote>")
	}
	return ast.WalkContinue, nil
}

func (r *customHTMLRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.FencedCodeBlock)
	language := string(n.Language(source))

	var codeBuf bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		codeBuf.Write(line.Value(source))
	}

	codeStr := codeBuf.String()
	highlighted := highlightCode(codeStr, language)

	fmt.Fprintf(w, `<div class="code-block-wrapper"><div class="code-block-header"><span class="code-block-lang">%s</span><button type="button" class="code-copy-btn" onclick="navigator.clipboard.writeText(this.closest('.code-block-wrapper').querySelector('code').innerText);this.innerText='Copied!';setTimeout(()=>this.innerText='Copy',2000)">Copy</button></div>%s</div>`,
		template.HTMLEscapeString(getLangDisplay(language)), highlighted)

	return ast.WalkContinue, nil
}

func (r *customHTMLRenderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.CodeBlock)
	var codeBuf bytes.Buffer
	l := n.Lines().Len()
	for i := 0; i < l; i++ {
		line := n.Lines().At(i)
		codeBuf.Write(line.Value(source))
	}

	codeStr := codeBuf.String()
	highlighted := highlightCode(codeStr, "")

	fmt.Fprintf(w, `<div class="code-block-wrapper"><div class="code-block-header"><span class="code-block-lang">text</span></div>%s</div>`, highlighted)

	return ast.WalkContinue, nil
}

func getLangDisplay(lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	if lang == "" {
		return "code"
	}
	return lang
}

func highlightCode(code, lang string) string {
	lang = strings.TrimSpace(strings.ToLower(lang))
	var lexer chroma.Lexer
	if lang != "" {
		lexer = lexers.Get(lang)
	}
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	formatter := chromahtml.New(
		chromahtml.WithClasses(true),
		chromahtml.PreventSurroundingPre(false),
	)

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return fmt.Sprintf("<pre><code>%s</code></pre>", template.HTMLEscapeString(code))
	}

	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return fmt.Sprintf("<pre><code>%s</code></pre>", template.HTMLEscapeString(code))
	}

	return buf.String()
}

type customExtension struct{}

func (e *customExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newCustomHTMLRenderer(), 100),
		),
	)
}

var defaultMarkdown = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
		extension.Linkify,
		&customExtension{},
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

// Render converts markdown text to sanitized/safe HTML wrapped with context styling classes.
func Render(textInput string, contextType ContextType) template.HTML {
	trimmed := strings.TrimSpace(textInput)
	if trimmed == "" {
		return ""
	}

	processed := preprocessMarkdown(trimmed)

	var buf bytes.Buffer
	if err := defaultMarkdown.Convert([]byte(processed), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(trimmed))
	}

	renderedHTML := postprocessHTML(buf.String())

	wrapperClass := "markdown-body"
	switch contextType {
	case ContextMessage:
		wrapperClass = "markdown-body markdown-message"
	case ContextChannelDescription:
		wrapperClass = "markdown-body markdown-channel-desc"
	case ContextGroupDescription:
		wrapperClass = "markdown-body markdown-group-desc"
	}

	htmlResult := fmt.Sprintf(`<div class="%s">%s</div>`, wrapperClass, renderedHTML)
	return template.HTML(htmlResult)
}
