package command

import (
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
)

// resolveNodeDocs はグラフの各ノードに対応する content.Doc を ID → Doc で返す (未解決ノードは含まない)。
// exclude のノード (公開ビルドの下書き) は含めない。改版履歴に下書きの updated / changes を渡さないため。
func resolveNodeDocs(g *graph.Graph, docs map[string]*content.Doc, exclude map[string]bool) map[string]*content.Doc {
	out := make(map[string]*content.Doc, len(g.Nodes))
	for _, n := range g.Nodes {
		if exclude[n.ID] {
			continue
		}
		if doc, ok := lookupDoc(docs, n.Node); ok {
			out[n.ID] = doc
		}
	}
	return out
}
