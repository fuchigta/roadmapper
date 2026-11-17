package command

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"

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

func TestResolveAssetBase(t *testing.T) {
	tests := []struct {
		basePath string
		want     string
	}{
		{"", "../"},
		{"/", "/"},
		{"/my-repo", "/my-repo/"},
		{"/my-repo/", "/my-repo/"},
	}
	for _, tt := range tests {
		t.Run("basePath="+tt.basePath, func(t *testing.T) {
			if got := resolveAssetBase(tt.basePath); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRecentBadges(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	day := func(s string) time.Time {
		d, err := time.Parse("2006-01-02", s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	log := &changelog.Log{NodeUpdated: map[string]time.Time{
		"recent": day("2026-03-20"),
		"old":    day("2026-01-01"),
		"future": day("2026-04-10"),
	}}
	got := recentBadges(log, now)
	want := map[string]time.Time{"recent": day("2026-03-20")}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for id, d := range want {
		if !got[id].Equal(d) {
			t.Errorf("%s: got %v, want %v", id, got[id], d)
		}
	}
}

func TestWriteOptionalFile(t *testing.T) {
	genErr := errors.New("boom")
	tests := []struct {
		name      string
		data      string
		err       error
		wantFile  bool
		wantErrIn string
	}{
		{"内容あり", "<xml/>", nil, true, ""},
		{"空なら書かない", "", nil, false, ""},
		{"生成エラー", "", genErr, false, "feed.rss の生成に失敗"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			err := writeOptionalFile(dir, "feed.rss", func() (string, error) { return tt.data, tt.err })
			if tt.wantErrIn != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrIn) || !errors.Is(err, genErr) {
					t.Fatalf("err = %v, want wrapped error containing %q", err, tt.wantErrIn)
				}
			} else if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			got, readErr := os.ReadFile(filepath.Join(dir, "feed.rss"))
			if tt.wantFile {
				if readErr != nil || string(got) != tt.data {
					t.Errorf("file = %q (%v), want %q", got, readErr, tt.data)
				}
			} else if !os.IsNotExist(readErr) {
				t.Errorf("file should not exist: %v", readErr)
			}
		})
	}
}

func TestUnresolvedLabel(t *testing.T) {
	tests := []struct {
		name string
		node config.Node
		want string
	}{
		{"content なし", config.Node{ID: "a"}, "a"},
		{"content あり", config.Node{ID: "a", Content: "sub/page"}, `a (content: "sub/page")`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unresolvedLabel(&tt.node); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNodePlainText(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		links []config.Link
		want  string
	}{
		{"リンクなし", "本文", nil, "本文"},
		{"タイトルありリンク", "本文", []config.Link{{Title: "MDN", URL: "u1"}, {Title: "Spec", URL: "u2"}}, "本文 MDN Spec"},
		{"空タイトルは除く", "本文", []config.Link{{Title: "", URL: "u1"}, {Title: "MDN", URL: "u2"}}, "本文 MDN"},
		{"全リンクが空タイトル", "本文", []config.Link{{URL: "u1"}}, "本文"},
		{"本文が空", "", []config.Link{{Title: "MDN", URL: "u1"}}, "MDN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nodePlainText(tt.body, tt.links); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApplyFrontmatterLinks(t *testing.T) {
	orig := []config.Link{{Title: "yml", URL: "https://yml.example"}}
	tests := []struct {
		name    string
		fmLinks []content.Link
		want    []config.Link
	}{
		{"frontmatter なしなら roadmap.yml のまま", nil, orig},
		{"frontmatter が優先", []content.Link{{Title: "fm", URL: "https://fm.example"}},
			[]config.Link{{Title: "fm", URL: "https://fm.example"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &config.Node{ID: "a", Links: orig}
			applyFrontmatterLinks(n, tt.fmLinks)
			if len(n.Links) != len(tt.want) {
				t.Fatalf("got %v, want %v", n.Links, tt.want)
			}
			for i := range tt.want {
				if n.Links[i] != tt.want[i] {
					t.Errorf("links[%d] = %v, want %v", i, n.Links[i], tt.want[i])
				}
			}
		})
	}
}

func TestRenderNodeDoc(t *testing.T) {
	tests := []struct {
		name        string
		node        config.Node
		doc         content.Doc
		wantMermaid bool
		wantIn      []string
		wantNotIn   []string
	}{
		{
			name:   "相対画像はサブディレクトリ込みで書き換える",
			node:   config.Node{ID: "a"},
			doc:    content.Doc{RelDir: "sub", Body: "![x](img.png)"},
			wantIn: []string{`src="../content/sub/img.png"`},
		},
		{
			name:        "mermaid を検出する",
			node:        config.Node{ID: "a"},
			doc:         content.Doc{Body: "```mermaid\ngraph TD; A-->B;\n```\n"},
			wantMermaid: true,
		},
		{
			name:      "frontmatter のリンクを本文の後ろに追記する",
			node:      config.Node{ID: "a", Links: []config.Link{{Title: "yml", URL: "https://yml.example"}}},
			doc:       content.Doc{Body: "本文", Frontmatter: content.Frontmatter{Links: []content.Link{{Title: "fm", URL: "https://fm.example"}}}},
			wantIn:    []string{"本文", "https://fm.example"},
			wantNotIn: []string{"https://yml.example"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, mermaid, err := renderNodeDoc(&tt.node, &tt.doc, "../")
			if err != nil {
				t.Fatal(err)
			}
			if mermaid != tt.wantMermaid {
				t.Errorf("mermaid = %v, want %v", mermaid, tt.wantMermaid)
			}
			for _, s := range tt.wantIn {
				if !strings.Contains(html, s) {
					t.Errorf("html に %q がない: %s", s, html)
				}
			}
			for _, s := range tt.wantNotIn {
				if strings.Contains(html, s) {
					t.Errorf("html に %q がある: %s", s, html)
				}
			}
		})
	}
}
