// Package render はロードマップを SVG や HTML に変換する。
package render

import (
	"fmt"
	"strings"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
)

const fontSize = 13

const svgPadding = 40.0

// EdgeStyle はエッジの見た目を決める。
type EdgeStyle struct {
	Color     string
	Dash      string
	ArrowHead string
}

var edgeStyles = map[config.NodeType]EdgeStyle{
	config.NodeTypeRequired:    {Color: "#374151", Dash: "none", ArrowHead: "filled"},
	config.NodeTypeOptional:    {Color: "#9ca3af", Dash: "6 3", ArrowHead: "open"},
	config.NodeTypeAlternative: {Color: "#6366f1", Dash: "4 4", ArrowHead: "open"},
}

// nodeColor は進捗状態に関わらずデフォルトのノード色を type 別に返す。
var nodeColor = map[config.NodeType]struct{ fill, stroke, text string }{
	config.NodeTypeRequired:    {"#1e293b", "#1e293b", "#f8fafc"},
	config.NodeTypeOptional:    {"#f8fafc", "#9ca3af", "#374151"},
	config.NodeTypeAlternative: {"#ede9fe", "#6366f1", "#4338ca"},
}

// RenderSVG は graph g を SVG 文字列に変換する。
func RenderSVG(g *graph.Graph, lr *layout.Result, brandColor string) string {
	w := lr.Width + svgPadding*2
	h := lr.Height + svgPadding*2

	var sb strings.Builder

	// SVG ヘッダ
	fmt.Fprintf(&sb,
		`<svg xmlns="http://www.w3.org/2000/svg" `+
			`width="%v" height="%v" `+
			`viewBox="0 0 %v %v" `+
			`class="roadmap-svg">`,
		w, h, w, h)

	// defs (マーカー)
	sb.WriteString(buildDefs())

	// ルートグループ (padding offset)
	fmt.Fprintf(&sb, `<g transform="translate(%v,%v)">`, svgPadding, svgPadding)

	// エッジを先に描く (ノードの下になるように)
	for _, n := range g.Nodes {
		for _, child := range n.ChildrenNodes {
			renderEdge(&sb, n, child, lr)
		}
	}

	// ノードを描く
	for _, n := range g.Nodes {
		renderNode(&sb, n, lr)
	}

	sb.WriteString(`</g></svg>`)
	return sb.String()
}

func buildDefs() string {
	return `<defs>
  <marker id="arrow-required" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
    <polygon points="0 0, 10 3.5, 0 7" fill="#374151"/>
  </marker>
  <marker id="arrow-optional" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
    <polygon points="0 0, 10 3.5, 0 7" fill="#9ca3af"/>
  </marker>
  <marker id="arrow-alternative" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
    <polygon points="0 0, 10 3.5, 0 7" fill="#6366f1"/>
  </marker>
  <filter id="shadow" x="-5%" y="-5%" width="110%" height="110%">
    <feDropShadow dx="0" dy="2" stdDeviation="3" flood-color="#00000020"/>
  </filter>
</defs>`
}

func renderNode(sb *strings.Builder, n *graph.Node, lr *layout.Result) {
	nl, ok := lr.Nodes[n.ID]
	if !ok {
		return
	}

	x := nl.X - nl.Width/2
	y := nl.Y - nl.Height/2
	rx := 8.0

	colors := nodeColor[n.Node.Type]

	// 親子IDリスト
	parentIDs := make([]string, len(n.ParentNodes))
	for i, p := range n.ParentNodes {
		parentIDs[i] = p.ID
	}
	childIDs := make([]string, len(n.ChildrenNodes))
	for i, c := range n.ChildrenNodes {
		childIDs[i] = c.ID
	}

	// ノード外枠 (クリッカブル要素)
	fmt.Fprintf(sb,
		`<g class="roadmap-node" data-id=%q data-type=%q `+
			`data-parents=%q data-children=%q `+
			`transform="translate(%v,%v)" `+
			`style="cursor:pointer">`,
		n.ID, string(n.Node.Type),
		strings.Join(parentIDs, ","), strings.Join(childIDs, ","),
		0, 0)

	// 影付き矩形
	fmt.Fprintf(sb,
		`<rect x="%v" y="%v" width="%v" height="%v" rx="%v" `+
			`fill="%s" stroke="%s" stroke-width="1.5" `+
			`filter="url(#shadow)"/>`,
		x, y, nl.Width, nl.Height, rx,
		colors.fill, colors.stroke)

	// テキスト (中央揃え・複数行対応)
	maxInner := nl.Width - layout.NodePaddingX
	lines := layout.WrapTitle(n.Title, fontSize, maxInner)
	nLines := len(lines)
	// 1 行目の y オフセット: 行ブロック全体を縦中央に揃える
	firstDY := -float64(nLines-1) * layout.LineHeight / 2
	fmt.Fprintf(sb,
		`<text x="%v" y="%v" `+
			`text-anchor="middle" dominant-baseline="middle" `+
			`font-family="system-ui,sans-serif" font-size="%d" `+
			`fill="%s" font-weight="500">`,
		nl.X, nl.Y, fontSize, colors.text)
	for i, line := range lines {
		if i == 0 {
			fmt.Fprintf(sb, `<tspan x="%v" dy="%.4g">%s</tspan>`, nl.X, firstDY, escapeXML(line))
		} else {
			fmt.Fprintf(sb, `<tspan x="%v" dy="%.4g">%s</tspan>`, nl.X, layout.LineHeight, escapeXML(line))
		}
	}
	sb.WriteString(`</text>`)

	// type バッジ (optional / alternative のみ)
	if n.Node.Type != config.NodeTypeRequired {
		badge := "opt"
		badgeColor := "#9ca3af"
		if n.Node.Type == config.NodeTypeAlternative {
			badge = "alt"
			badgeColor = "#6366f1"
		}
		fmt.Fprintf(sb,
			`<rect x="%v" y="%v" width="28" height="14" rx="7" fill="%s" opacity="0.9"/>`,
			x+nl.Width-30, y+1, badgeColor)
		fmt.Fprintf(sb,
			`<text x="%v" y="%v" text-anchor="middle" dominant-baseline="middle" `+
				`font-family="system-ui,sans-serif" font-size="8" fill="white">%s</text>`,
			x+nl.Width-16, y+8, badge)
	}

	// difficulty バッジ (左上)
	if n.Node.Difficulty != "" {
		var diffLabel, diffColor string
		switch n.Node.Difficulty {
		case config.DifficultyBeginner:
			diffLabel, diffColor = "初", "#22c55e"
		case config.DifficultyIntermediate:
			diffLabel, diffColor = "中", "#f59e0b"
		case config.DifficultyAdvanced:
			diffLabel, diffColor = "上", "#ef4444"
		}
		fmt.Fprintf(sb,
			`<rect x="%v" y="%v" width="18" height="14" rx="7" fill="%s" opacity="0.9"/>`,
			x+2, y+1, diffColor)
		fmt.Fprintf(sb,
			`<text x="%v" y="%v" text-anchor="middle" dominant-baseline="middle" `+
				`font-family="system-ui,sans-serif" font-size="8" fill="white">%s</text>`,
			x+11, y+8, diffLabel)
	}

	sb.WriteString(`</g>`)
}

func renderEdge(sb *strings.Builder, parent, child *graph.Node, lr *layout.Result) {
	pnl, ok1 := lr.Nodes[parent.ID]
	cnl, ok2 := lr.Nodes[child.ID]
	if !ok1 || !ok2 {
		return
	}

	// エッジの type は子ノードの type に従う
	style := edgeStyles[child.Node.Type]

	// 親の下端 → 子の上端 への折れ線パス
	x1 := pnl.X
	y1 := pnl.Y + pnl.Height/2
	x2 := cnl.X
	y2 := cnl.Y - cnl.Height/2

	midY := (y1 + y2) / 2

	markerID := "arrow-" + string(child.Node.Type)

	var dashAttr string
	if style.Dash != "none" {
		dashAttr = fmt.Sprintf(`stroke-dasharray="%s"`, style.Dash)
	}

	fmt.Fprintf(sb,
		`<path d="M %v %v C %v %v %v %v %v %v" `+
			`fill="none" stroke="%s" stroke-width="1.5" %s `+
			`marker-end="url(#%s)" `+
			`data-source=%q data-target=%q class="roadmap-edge"/>`,
		x1, y1,
		x1, midY, x2, midY,
		x2, y2,
		style.Color, dashAttr, markerID,
		parent.ID, child.ID)
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
