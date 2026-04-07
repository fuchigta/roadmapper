package render

import (
	"bytes"
	"regexp"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"
)

// checkboxDisabledRe は goldmark が出力する input タグ内の disabled="" 属性にマッチする。
var checkboxDisabledRe = regexp.MustCompile(` disabled=""`)

// mermaidBlockRe は goldmark-highlighting が出力する mermaid コードブロックにマッチする。
// 例: <pre tabindex="0" ...><code class="language-mermaid">...</code></pre>
// または chroma のラップなし <pre><code class="language-mermaid">
var mermaidBlockRe = regexp.MustCompile(
	`(?s)<pre[^>]*><code[^>]*class="[^"]*language-mermaid[^"]*"[^>]*>(.*?)</code></pre>`,
)

// RenderMarkdown は Markdown 文字列を HTML に変換する。
// mermaid コードブロックは <pre class="mermaid"> でパススルーする。
func RenderMarkdown(src string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Typographer,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithFormatOptions(
					html.WithClasses(false),
					html.WithLineNumbers(false),
				),
				highlighting.WithGuessLanguage(false),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			goldmarkHTML.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(src), &buf); err != nil {
		return "", err
	}

	result := buf.String()

	// mermaid コードブロックを <pre class="mermaid"> に変換
	result = mermaidBlockRe.ReplaceAllStringFunc(result, func(match string) string {
		subs := mermaidBlockRe.FindStringSubmatch(match)
		if len(subs) < 2 {
			return match
		}
		return `<pre class="mermaid">` + unescapeHTML(subs[1]) + `</pre>`
	})

	// チェックリストの disabled="" 属性を除去して app.js から操作できるようにする
	result = checkboxDisabledRe.ReplaceAllString(result, ``)

	return result, nil
}

// unescapeHTML は &amp; &lt; &gt; &quot; を戻す (mermaid ソース復元用)。
func unescapeHTML(s string) string {
	s = regexp.MustCompile(`&amp;`).ReplaceAllString(s, "&")
	s = regexp.MustCompile(`&lt;`).ReplaceAllString(s, "<")
	s = regexp.MustCompile(`&gt;`).ReplaceAllString(s, ">")
	s = regexp.MustCompile(`&quot;`).ReplaceAllString(s, `"`)
	return s
}

// ChromaCSS はシンタックスハイライト用の CSS を返す。
func ChromaCSS() string {
	style := styles.Get("github")
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(html.WithClasses(true))
	var buf bytes.Buffer
	_ = formatter.WriteCSS(&buf, style)
	return buf.String()
}

// mermaidRe は HTML 文字列に mermaid ブロックが含まれるか判定する。
var mermaidRe = regexp.MustCompile(`<pre class="mermaid">`)

// HasMermaid は html 文字列に mermaid ブロックが含まれるか返す。
func HasMermaid(html string) bool {
	return mermaidRe.MatchString(html)
}
