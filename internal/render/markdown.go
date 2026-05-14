package render

import (
	"bytes"
	"path"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
)

// checkboxDisabledRe は goldmark が出力する input タグ内の disabled="" 属性にマッチする。
var checkboxDisabledRe = regexp.MustCompile(` disabled=""`)

// mermaidBlockRe は goldmark-highlighting が出力する mermaid コードブロックにマッチする。
// 例: <pre tabindex="0" ...><code class="language-mermaid">...</code></pre>
// または chroma のラップなし <pre><code class="language-mermaid">
var mermaidBlockRe = regexp.MustCompile(
	`(?s)<pre[^>]*><code[^>]*class="[^"]*language-mermaid[^"]*"[^>]*>(.*?)</code></pre>`,
)

// md はパッケージ全体で共有する goldmark インスタンス。
var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Typographer,
		highlighting.NewHighlighting(
			highlighting.WithStyle("github"),
			highlighting.WithFormatOptions(
				html.WithClasses(true),
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

// RenderMarkdown は Markdown 文字列を HTML に変換する。
// mermaid コードブロックは <pre class="mermaid"> でパススルーする。
func RenderMarkdown(src string) (string, error) {
	return RenderMarkdownWithBase(src, "")
}

// RenderMarkdownWithBase は Markdown を HTML に変換しつつ、
// 画像・リンクの相対 URL に urlPrefix を付加する。
//
// urlPrefix が空の場合は URL を変更しない。
// スキーム付き URL (http://, mailto: など)、ルート相対パス (/foo)、
// フラグメントのみ (#section) はそのまま保持する。
func RenderMarkdownWithBase(src, urlPrefix string) (string, error) {
	source := []byte(src)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	if urlPrefix != "" {
		rewriteRelativeURLs(doc, urlPrefix)
	}

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, source, doc); err != nil {
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

var (
	reAmp  = regexp.MustCompile(`&amp;`)
	reLt   = regexp.MustCompile(`&lt;`)
	reGt   = regexp.MustCompile(`&gt;`)
	reQuot = regexp.MustCompile(`&quot;`)
)

// unescapeHTML は &amp; &lt; &gt; &quot; を戻す (mermaid ソース復元用)。
func unescapeHTML(s string) string {
	s = reAmp.ReplaceAllString(s, "&")
	s = reLt.ReplaceAllString(s, "<")
	s = reGt.ReplaceAllString(s, ">")
	s = reQuot.ReplaceAllString(s, `"`)
	return s
}

// rewriteRelativeURLs は AST 内の画像・リンクの相対 URL に prefix を付加する。
func rewriteRelativeURLs(doc ast.Node, prefix string) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Image:
			v.Destination = rewriteDestination(v.Destination, prefix)
		case *ast.Link:
			v.Destination = rewriteDestination(v.Destination, prefix)
		}
		return ast.WalkContinue, nil
	})
}

func rewriteDestination(dest []byte, prefix string) []byte {
	s := string(dest)
	if s == "" || isAbsoluteURL(s) {
		return dest
	}
	// クエリ/フラグメントを保持
	suffix := ""
	if i := strings.IndexAny(s, "?#"); i >= 0 {
		suffix = s[i:]
		s = s[:i]
	}
	if s == "" {
		return dest
	}
	cleaned := path.Clean(prefix + s)
	return []byte(cleaned + suffix)
}

// isAbsoluteURL は dest が書き換え不要な絶対 URL/パス/フラグメントか判定する。
func isAbsoluteURL(s string) bool {
	if strings.HasPrefix(s, "/") { // ルート相対 or "//" プロトコル相対
		return true
	}
	if strings.HasPrefix(s, "#") { // フラグメントのみ
		return true
	}
	// scheme:... 形式 (http, https, mailto, data, tel, ftp など)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ':' {
			return i > 0
		}
		if !isSchemeChar(c) {
			return false
		}
	}
	return false
}

func isSchemeChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '+' || c == '-' || c == '.'
}

// ChromaCSS はシンタックスハイライト用の CSS を返す (ライト + ダーク)。
func ChromaCSS() string {
	light := chromaCSSForStyle("github", "")
	dark := chromaCSSForStyle("dracula", "[data-theme=\"dark\"] ")
	return light + "\n" + dark
}

// chromaCSSForStyle は指定スタイルの Chroma CSS を生成し、各ルールに prefix を付ける。
func chromaCSSForStyle(styleName, prefix string) string {
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	formatter := html.New(html.WithClasses(true))
	var buf bytes.Buffer
	_ = formatter.WriteCSS(&buf, style)
	if prefix == "" {
		return buf.String()
	}
	return strings.ReplaceAll(buf.String(), ".chroma", prefix+".chroma")
}

// mermaidRe は HTML 文字列に mermaid ブロックが含まれるか判定する。
var mermaidRe = regexp.MustCompile(`<pre class="mermaid">`)

// HasMermaid は html 文字列に mermaid ブロックが含まれるか返す。
func HasMermaid(html string) bool {
	return mermaidRe.MatchString(html)
}

// ExtractPlainText は Markdown ソースからプレーンテキストを抽出する。
// 見出し・本文・リスト・コードブロック内容を空白区切りで返す。
// 全文検索インデックス用。HTML タグやエンティティは出現しない。
func ExtractPlainText(src string) string {
	source := []byte(src)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var sb strings.Builder
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Text:
			sb.Write(v.Segment.Value(source))
			sb.WriteByte(' ')
		case *ast.CodeBlock:
			for i := 0; i < v.Lines().Len(); i++ {
				seg := v.Lines().At(i)
				sb.Write(seg.Value(source))
				sb.WriteByte(' ')
			}
		case *ast.FencedCodeBlock:
			for i := 0; i < v.Lines().Len(); i++ {
				seg := v.Lines().At(i)
				sb.Write(seg.Value(source))
				sb.WriteByte(' ')
			}
		case *ast.AutoLink:
			sb.Write(v.Label(source))
			sb.WriteByte(' ')
		}
		return ast.WalkContinue, nil
	})

	return strings.Join(strings.Fields(sb.String()), " ")
}
