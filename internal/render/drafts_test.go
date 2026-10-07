package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/render"
	"github.com/fuchigta/roadmapper/web"
)

const (
	secretBody = "SECRET-BODY-9f3a"
	secretLink = "https://secret.example.com/draft-link"
	secretEdit = "docs/content/secret-draft.md"
)

// draftFixture は a → d(下書き) → c の 3 ノード構成。
func draftFixture(t *testing.T) (*config.Config, *graph.Graph, *layout.Result) {
	t.Helper()
	cfg := &config.Config{
		Site: config.Site{
			BrandColor: "#4f46e5",
			Layout:     config.Layout{RankDir: "TB", NodeSep: 50, RankSep: 80},
		},
		Roadmaps: []config.Roadmap{{
			ID: "test", Title: "Test Roadmap",
			Nodes: []*config.Node{
				{ID: "a", Title: "Node A", Type: config.NodeTypeRequired, Children: []*config.Node{
					{ID: "d", Title: "Draft D", Type: config.NodeTypeRequired, Draft: true,
						Links:    []config.Link{{Title: "SECRET-LINK-TITLE", URL: secretLink}},
						Children: []*config.Node{{ID: "c", Title: "Node C", Type: config.NodeTypeOptional}}},
				}},
			},
		}},
	}
	g, err := graph.Build(&cfg.Roadmaps[0])
	if err != nil {
		t.Fatalf("graph.Build: %v", err)
	}
	lr, err := layout.Compute(g, cfg)
	if err != nil {
		t.Fatalf("layout.Compute: %v", err)
	}
	return cfg, g, lr
}

func renderDraftPage(t *testing.T, drafts render.Drafts) string {
	t.Helper()
	cfg, g, lr := draftFixture(t)
	nodeHTML := map[string]string{"a": "<p>A body</p>", "d": "<p>" + secretBody + "</p>"}
	nodeText := map[string]string{"a": "A body", "d": secretBody}
	editPaths := map[string]string{"a": "docs/content/a.md", "d": secretEdit}
	html, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nodeHTML, nodeText, "/", "../", false, nil, nil, editPaths, drafts)
	if err != nil {
		t.Fatalf("RenderRoadmapPage: %v", err)
	}
	return strings.Join(strings.Fields(html), " ")
}

// 公開ビルドでは下書きノードの本文・リンク・編集パスが HTML のどこにも出ない。
func TestRenderRoadmapPage_draftLocked_noLeak(t *testing.T) {
	html := renderDraftPage(t, render.Drafts{IDs: map[string]bool{"d": true}})
	for _, s := range []string{secretBody, secretLink, "SECRET-LINK-TITLE", secretEdit} {
		if strings.Contains(html, s) {
			t.Errorf("下書きの内容 %q が出力に含まれています", s)
		}
	}
	for _, want := range []string{
		`"draft":true`, `class="roadmap-node is-draft"`, `data-draft="1"`, `aria-disabled="true"`,
		`<title>準備中</title>`, `draftPreview: false`, `A body`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("出力に %q が含まれていません", want)
		}
	}
	if !strings.Contains(html, `"editPath":"docs/content/a.md"`) {
		t.Error("通常ノードの editPath は残るはず")
	}
}

// プレビューでは本文を含めた上で draft フラグと draftPreview を渡す。
func TestRenderRoadmapPage_draftPreview(t *testing.T) {
	html := renderDraftPage(t, render.Drafts{IDs: map[string]bool{"d": true}, Preview: true})
	for _, want := range []string{
		secretBody, secretEdit, `"draft":true`, `draftPreview: true`,
		`class="roadmap-node is-draft-preview"`, `node-draft-badge`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("出力に %q が含まれていません", want)
		}
	}
	if strings.Contains(html, `class="roadmap-node is-draft"`) || strings.Contains(html, "準備中") {
		t.Error("プレビューでは非活性表示にしない")
	}
}

// 下書きなしの出力には draft 関連の印が出ない。
func TestRenderRoadmapPage_noDrafts(t *testing.T) {
	html := renderDraftPage(t, render.Drafts{})
	for _, s := range []string{`"draft"`, "is-draft", "準備中", "node-draft-badge"} {
		if strings.Contains(html, s) {
			t.Errorf("下書きなしなのに %q が含まれています", s)
		}
	}
}

func TestRenderIndexPage_draftExcluded(t *testing.T) {
	cfg, g, _ := draftFixture(t)
	graphs := map[string]*graph.Graph{"test": g}
	tests := []struct {
		name   string
		locked map[string]map[string]bool
		want   string
	}{
		{"除外なし", nil, `"test":["a","d"]`},
		{"d を除外", map[string]map[string]bool{"test": {"d": true}}, `"test":["a"]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := render.RenderIndexPage(web.FS, cfg, "/", graphs, nil, tt.locked)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(html, tt.want) {
				t.Errorf("出力に %s が含まれていません", tt.want)
			}
		})
	}
}

// SVG のゴールデン比較 (UPDATE_GOLDEN=1 で更新)。
func TestRenderSVG_draftGolden(t *testing.T) {
	cfg, g, lr := draftFixture(t)
	tests := []struct {
		name   string
		drafts render.Drafts
	}{
		{"locked", render.Drafts{IDs: map[string]bool{"d": true}}},
		{"preview", render.Drafts{IDs: map[string]bool{"d": true}, Preview: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := render.RenderSVGWithBadges(g, lr, cfg.Site.BrandColor, nil, tt.drafts)
			path := filepath.Join("testdata", "svg_draft_"+tt.name+".svg")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ゴールデンを読めません (UPDATE_GOLDEN=1 で生成): %v", err)
			}
			if got != strings.ReplaceAll(string(want), "\r\n", "\n") {
				t.Errorf("SVG が %s と一致しません", path)
			}
		})
	}
}

// 下書きに接続するエッジだけが is-draft になる。
func TestRenderSVG_draftEdges(t *testing.T) {
	cfg, g, lr := draftFixture(t)
	got := render.RenderSVGWithBadges(g, lr, cfg.Site.BrandColor, nil, render.Drafts{IDs: map[string]bool{"c": true}})
	if n := strings.Count(got, `class="roadmap-edge is-draft"`); n != 1 {
		t.Errorf("is-draft エッジ数 = %d, want 1 (d→c のみ)", n)
	}
	if n := strings.Count(got, `class="roadmap-edge"`); n != 1 {
		t.Errorf("通常エッジ数 = %d, want 1 (a→d)", n)
	}
}
