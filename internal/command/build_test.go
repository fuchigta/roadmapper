package command

import (
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
)

func TestBuildEditPaths(t *testing.T) {
	rm := &config.Roadmap{
		ID: "r",
		Nodes: []*config.Node{
			{ID: "html", Title: "HTML", Type: config.NodeTypeRequired},
			{ID: "x", Title: "X", Type: config.NodeTypeRequired, Content: "frontend/css"},
			{ID: "none", Title: "None", Type: config.NodeTypeRequired},
			{ID: "nocontent", Title: "NC", Type: config.NodeTypeRequired, Content: "missing/page"},
		},
	}
	g, err := graph.Build(rm)
	if err != nil {
		t.Fatal(err)
	}
	html := &content.Doc{ID: "frontend/html", RelPath: "frontend/html.md"}
	css := &content.Doc{ID: "frontend/css", RelPath: "frontend/css.md"}
	docs := map[string]*content.Doc{"frontend/html": html, "html": html, "frontend/css": css, "css": css}

	tests := []struct {
		prefix string
		want   map[string]string
	}{
		{"", map[string]string{
			"html": "content/frontend/html.md", "x": "content/frontend/css.md",
			"none": "content/none.md", "nocontent": "content/missing/page.md"}},
		{"docs/", map[string]string{
			"html": "docs/content/frontend/html.md", "x": "docs/content/frontend/css.md",
			"none": "docs/content/none.md", "nocontent": "docs/content/missing/page.md"}},
	}
	for _, tt := range tests {
		t.Run("prefix="+tt.prefix, func(t *testing.T) {
			got := buildEditPaths(g, docs, tt.prefix)
			for id, want := range tt.want {
				if got[id] != want {
					t.Errorf("%s: got %q, want %q", id, got[id], want)
				}
			}
		})
	}
}
