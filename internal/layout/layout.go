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
)

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
		fmt.Fprintf(&sb, "g.setNode(%q, {label: %q, width: %v, height: %v});\n",
			n.ID, n.Title, defaultNodeWidth, defaultNodeHeight)
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
