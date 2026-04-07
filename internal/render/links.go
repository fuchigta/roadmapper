package render

import (
	"fmt"
	"strings"

	"github.com/fuchigta/roadmapper/internal/config"
)

// RenderLinks はリンク集の HTML フラグメントを生成する。
func RenderLinks(links []config.Link) string {
	if len(links) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<div class="node-links"><h3>参考資料</h3><ul>`)
	for _, l := range links {
		title := l.Title
		if title == "" {
			title = l.URL
		}
		fmt.Fprintf(&sb,
			`<li><a href="%s" target="_blank" rel="noopener noreferrer">%s</a></li>`,
			escapeXML(l.URL), escapeXML(title),
		)
	}
	sb.WriteString("</ul></div>")
	return sb.String()
}
