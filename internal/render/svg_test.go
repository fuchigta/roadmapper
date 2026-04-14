package render_test

import (
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/render"
)

func buildSVGWithNode(t *testing.T, node *config.Node) string {
	t.Helper()
	rm := &config.Roadmap{
		ID: "test", Title: "Test",
		Nodes: []*config.Node{node},
	}
	g, err := graph.Build(rm)
	if err != nil {
		t.Fatalf("graph.Build: %v", err)
	}
	cfg := &config.Config{
		Site: config.Site{
			BrandColor: "#4f46e5",
			Layout: config.Layout{
				RankDir: "TB",
				NodeSep: 50,
				RankSep: 80,
			},
		},
	}
	result, err := layout.Compute(g, cfg)
	if err != nil {
		t.Fatalf("layout.Compute: %v", err)
	}
	return render.RenderSVG(g, result, cfg.Site.BrandColor)
}

func buildSVG(t *testing.T, title string) string {
	t.Helper()
	return buildSVGWithNode(t, &config.Node{ID: "a", Title: title, Type: config.NodeTypeRequired})
}

func TestRenderSVG_shortTitle_singleLine(t *testing.T) {
	svg := buildSVG(t, "Go")
	count := strings.Count(svg, "<tspan")
	if count != 1 {
		t.Errorf("short title: expected 1 tspan, got %d", count)
	}
}

func TestRenderSVG_longTitle_multiLine(t *testing.T) {
	svg := buildSVG(t, "フロントエンド開発の基礎と応用について学ぶ")
	count := strings.Count(svg, "<tspan")
	if count < 2 {
		t.Errorf("long title: expected >= 2 tspans (got %d) — text should wrap", count)
	}
}

func TestRenderSVG_difficultyBadge(t *testing.T) {
	tests := []struct {
		name       string
		difficulty config.Difficulty
		wantLabel  string
	}{
		{"beginner", config.DifficultyBeginner, "初"},
		{"intermediate", config.DifficultyIntermediate, "中"},
		{"advanced", config.DifficultyAdvanced, "上"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg := buildSVGWithNode(t, &config.Node{
				ID:         "a",
				Title:      "Test",
				Type:       config.NodeTypeRequired,
				Difficulty: tt.difficulty,
			})
			if !strings.Contains(svg, tt.wantLabel) {
				t.Errorf("SVG should contain difficulty label %q", tt.wantLabel)
			}
		})
	}
}

func TestRenderSVG_noDifficultyBadge(t *testing.T) {
	svg := buildSVGWithNode(t, &config.Node{
		ID:    "a",
		Title: "Test",
		Type:  config.NodeTypeRequired,
	})
	for _, label := range []string{"初", "中", "上"} {
		if strings.Contains(svg, label) {
			t.Errorf("SVG should not contain difficulty label %q when difficulty is empty", label)
		}
	}
}
