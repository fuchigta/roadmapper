// Package layout は Goja + dagre.js を使ってグラフのノード座標を計算する。
package layout

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dop251/goja"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
)

//go:embed vendor/dagre.min.js
var dagreJS string

const (
	defaultNodeWidth  = 180.0
	defaultNodeHeight = 50.0
	minNodeWidth      = 180.0
	maxNodeWidth      = 260.0
	// NodePaddingX はノード左右の内側余白 (片側 12px × 2)。
	NodePaddingX = 24.0
	// NodePaddingY はノード上下の内側余白 (片側 10px × 2)。
	NodePaddingY = 20.0
	// LineHeight は複数行テキストの行送り (px)。
	LineHeight   = 18.0
	baseFontSize = 13.0
	maxLines     = 3
)

// measureTextWidth は title を baseFontSize で描画したときの推定横幅を返す。
// ASCII 半角: fontSize × 0.6、それ以外 (全角/CJK 等): fontSize × 1.05。
func measureTextWidth(s string, fontSize float64) float64 {
	var w float64
	for _, r := range s {
		if r < 128 {
			w += fontSize * 0.6
		} else {
			w += fontSize * 1.05
		}
	}
	return w
}

// WrapTitle は title を fontSize / maxInner px 幅に収まるよう折り返した行スライスを返す。
// 最大 maxLines 行まで折り返し、それを超える場合は最終行末尾を … で省略する。
// 空白を含む場合は単語境界、含まない場合はルーン単位で折り返す。
func WrapTitle(title string, fontSize, maxInner float64) []string {
	// 既に 1 行に収まる場合はそのまま返す
	if measureTextWidth(title, fontSize) <= maxInner {
		return []string{title}
	}

	words := strings.Fields(title)
	if len(words) > 1 {
		return wrapByWords(words, fontSize, maxInner)
	}
	return wrapByRunes([]rune(title), fontSize, maxInner)
}

func wrapByWords(words []string, fontSize, maxInner float64) []string {
	var lines []string
	current := ""
	for _, w := range words {
		candidate := w
		if current != "" {
			candidate = current + " " + w
		}
		if measureTextWidth(candidate, fontSize) <= maxInner {
			current = candidate
		} else {
			if current != "" {
				lines = append(lines, current)
				if len(lines) == maxLines-1 {
					// 残りを最終行にまとめて省略
					break
				}
			}
			current = w
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return truncateLines(lines, fontSize, maxInner)
}

func wrapByRunes(runes []rune, fontSize, maxInner float64) []string {
	var lines []string
	start := 0
	for start < len(runes) {
		end := start
		var w float64
		for end < len(runes) {
			charW := fontSize * 0.6
			if runes[end] >= 128 {
				charW = fontSize * 1.05
			}
			if w+charW > maxInner {
				break
			}
			w += charW
			end++
		}
		if end == start {
			end = start + 1 // 最低 1 文字は進める
		}
		lines = append(lines, string(runes[start:end]))
		start = end
		if len(lines) == maxLines-1 && start < len(runes) {
			// 残りを最終行候補に追加してから省略処理へ
			lines = append(lines, string(runes[start:]))
			break
		}
	}
	return truncateLines(lines, fontSize, maxInner)
}

// truncateLines は lines が maxLines を超えている場合、最終行を … で省略する。
func truncateLines(lines []string, fontSize, maxInner float64) []string {
	if len(lines) <= maxLines {
		// 最終行が幅を超えていたら省略
		last := lines[len(lines)-1]
		if measureTextWidth(last, fontSize) > maxInner {
			lines[len(lines)-1] = trimToWidth([]rune(last), fontSize, maxInner)
		}
		return lines
	}
	// maxLines に切り詰めて最終行を省略
	lines = lines[:maxLines]
	lines[maxLines-1] = trimToWidth([]rune(lines[maxLines-1]), fontSize, maxInner)
	return lines
}

// trimToWidth はルーン列を maxInner 以内に収まるよう末尾を … で省略する。
func trimToWidth(runes []rune, fontSize, maxInner float64) string {
	ellipsisW := fontSize * 0.6 // "…" は ASCII 幅相当で見積もる
	budget := maxInner - ellipsisW
	var w float64
	end := 0
	for end < len(runes) {
		charW := fontSize * 0.6
		if runes[end] >= 128 {
			charW = fontSize * 1.05
		}
		if w+charW > budget {
			break
		}
		w += charW
		end++
	}
	return string(runes[:end]) + "…"
}

// NodeSize は title に応じた最適なノード幅・高さを返す。
func NodeSize(title string) (width, height float64) {
	lines := WrapTitle(title, baseFontSize, maxNodeWidth-NodePaddingX)
	// 最長行の推定幅からノード幅を決定
	var maxLineW float64
	for _, l := range lines {
		if w := measureTextWidth(l, baseFontSize); w > maxLineW {
			maxLineW = w
		}
	}
	width = maxLineW + NodePaddingX
	if width < minNodeWidth {
		width = minNodeWidth
	}
	if width > maxNodeWidth {
		width = maxNodeWidth
	}
	height = float64(len(lines))*LineHeight + NodePaddingY
	if height < defaultNodeHeight {
		height = defaultNodeHeight
	}
	return width, height
}

// NodeLayout はレイアウト計算後の1ノードの座標情報。
type NodeLayout struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Result はグラフ全体のレイアウト結果。
type Result struct {
	Nodes  map[string]NodeLayout
	Width  float64
	Height float64
}

// Compute は g のレイアウトを dagre.js で計算して Result を返す。
// cfg.Site.Layout のパラメータを使う。
// ノードに x/y が手動指定されている場合はそちらで上書きする。
func Compute(g *graph.Graph, cfg *config.Config) (*Result, error) {
	vm := goja.New()

	if _, err := vm.RunString(dagreJS); err != nil {
		return nil, fmt.Errorf("dagre.js の初期化に失敗: %w", err)
	}

	layout := cfg.Site.Layout
	script := buildScript(g, layout)

	val, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("レイアウト計算に失敗: %w", err)
	}

	var raw map[string]map[string]float64
	if err := json.Unmarshal([]byte(val.String()), &raw); err != nil {
		return nil, fmt.Errorf("レイアウト結果のパースに失敗: %w", err)
	}

	result := &Result{Nodes: make(map[string]NodeLayout)}

	for id, pos := range raw {
		if id == "__graph__" {
			result.Width = pos["width"]
			result.Height = pos["height"]
			continue
		}
		nl := NodeLayout{
			X:      pos["x"],
			Y:      pos["y"],
			Width:  pos["width"],
			Height: pos["height"],
		}
		result.Nodes[id] = nl
	}

	// 手動座標オーバーライドを適用
	applyManualOverrides(g, result)

	return result, nil
}

func buildScript(g *graph.Graph, layout config.Layout) string {
	var sb strings.Builder

	sb.WriteString(`(function(){
var g = new dagre.graphlib.Graph();
g.setGraph({`)
	fmt.Fprintf(&sb, "rankdir: %q, nodesep: %v, ranksep: %v",
		layout.RankDir, layout.NodeSep, layout.RankSep)
	sb.WriteString(`});
g.setDefaultEdgeLabel(function(){ return {}; });
`)

	for _, n := range g.Nodes {
		w, h := NodeSize(n.Title)
		fmt.Fprintf(&sb, "g.setNode(%q, {label: %q, width: %v, height: %v});\n",
			n.ID, n.Title, w, h)
	}

	for _, n := range g.Nodes {
		for _, child := range n.ChildrenNodes {
			fmt.Fprintf(&sb, "g.setEdge(%q, %q);\n", n.ID, child.ID)
		}
	}

	sb.WriteString(`dagre.layout(g);
var result = {};
g.nodes().forEach(function(v){
  var nd = g.node(v);
  result[v] = {x: nd.x, y: nd.y, width: nd.width, height: nd.height};
});
var gi = g.graph();
result["__graph__"] = {x: 0, y: 0, width: gi.width, height: gi.height};
return JSON.stringify(result);
})()`)

	return sb.String()
}

func applyManualOverrides(g *graph.Graph, result *Result) {
	for _, n := range g.Nodes {
		if n.X == nil && n.Y == nil {
			continue
		}
		nl := result.Nodes[n.ID]
		if n.X != nil {
			nl.X = *n.X
		}
		if n.Y != nil {
			nl.Y = *n.Y
		}
		result.Nodes[n.ID] = nl
	}
}
