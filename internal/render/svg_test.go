package render_test

import (
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/render"
)

func buildSVG(t *testing.T, title string) string {
	t.Helper()
	rm := &config.Roadmap{
		ID: "test", Title: "Test",
		Nodes: []*config.Node{
			{ID: "a", Title: title, Type: config.NodeTypeRequired},
		},
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
