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
			got := buildEditPaths(g, docs, tt.prefix, nil)
			for id, want := range tt.want {
				if got[id] != want {
					t.Errorf("%s: got %q, want %q", id, got[id], want)
				}
			}
		})
	}
}

func TestResolveDrafts(t *testing.T) {
	rm := &config.Roadmap{
		ID: "r",
		Nodes: []*config.Node{
			{ID: "plain", Title: "P", Type: config.NodeTypeRequired},
			{ID: "fm", Title: "FM", Type: config.NodeTypeRequired},
			{ID: "yml", Title: "Y", Type: config.NodeTypeRequired, Draft: true},
			{ID: "both", Title: "B", Type: config.NodeTypeRequired, Draft: true},
			{ID: "noart", Title: "N", Type: config.NodeTypeRequired},
			{ID: "ymlnoart", Title: "YN", Type: config.NodeTypeRequired, Draft: true},
			{ID: "x", Title: "X", Type: config.NodeTypeRequired, Content: "sub/page"},
			{ID: "fmfalse", Title: "FF", Type: config.NodeTypeRequired},
		},
	}
	g, err := graph.Build(rm)
	if err != nil {
		t.Fatal(err)
	}
	doc := func(draft bool) *content.Doc {
		return &content.Doc{Frontmatter: content.Frontmatter{Draft: draft}}
	}
	docs := map[string]*content.Doc{
		"plain":    doc(false),
		"fm":       doc(true),
		"yml":      doc(false),
		"both":     doc(true),
		"sub/page": doc(true),
		"fmfalse":  doc(false),
	}
	got := resolveDrafts(g, docs)
	want := map[string]bool{"fm": true, "yml": true, "both": true, "ymlnoart": true, "x": true}
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	for _, n := range g.Nodes {
		if got[n.ID] != want[n.ID] {
			t.Errorf("%s: draft = %v, want %v", n.ID, got[n.ID], want[n.ID])
		}
	}
}

func TestDraftsLocked_helpers(t *testing.T) {
	rm := &config.Roadmap{ID: "r", Nodes: []*config.Node{
		{ID: "a", Title: "A", Type: config.NodeTypeRequired},
		{ID: "d", Title: "D", Type: config.NodeTypeRequired},
	}}
	g, err := graph.Build(rm)
	if err != nil {
		t.Fatal(err)
	}
	a := &content.Doc{RelPath: "a.md", Body: "公開本文", Frontmatter: content.Frontmatter{Updated: "2026-01-01"}}
	d := &content.Doc{RelPath: "d.md", Body: "SECRET-BODY", Frontmatter: content.Frontmatter{Draft: true, Updated: "2026-02-01"}}
	docs := map[string]*content.Doc{"a": a, "d": d}
	locked := resolveDrafts(g, docs)

	html, text, _, err := buildNodeHTML(g, docs, "../", locked)
	if err != nil {
		t.Fatal(err)
	}
	if html["d"] != "" || text["d"] != "" {
		t.Errorf("下書きの本文が出力されています: %q %q", html["d"], text["d"])
	}
	if html["a"] == "" || text["a"] == "" {
		t.Error("通常ノードの本文が出力されていません")
	}
	// プレビュー (locked = nil) では本文を出す
	if html, _, _, _ := buildNodeHTML(g, docs, "../", nil); html["d"] == "" {
		t.Error("プレビューでは下書きの本文も出力するはず")
	}

	if _, ok := buildEditPaths(g, docs, "", locked)["d"]; ok {
		t.Error("下書きの editPath が出力されています")
	}
	if buildEditPaths(g, docs, "", nil)["d"] != "content/d.md" {
		t.Error("プレビューでは下書きの editPath も出すはず")
	}

	if _, ok := resolveNodeDocs(g, docs, locked)["d"]; ok {
		t.Error("下書きの Doc が改版履歴の入力に含まれています")
	}
	if _, ok := resolveNodeDocs(g, docs, nil)["d"]; !ok {
		t.Error("exclude なしなら含まれるはず")
	}
}
