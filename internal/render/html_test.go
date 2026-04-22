package render_test

import (
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/render"
	"github.com/fuchigta/roadmapper/web"
)

func buildMinimalPageFixture(t *testing.T) (*config.Config, *graph.Graph, *layout.Result) {
	t.Helper()
	cfg := &config.Config{
		Site: config.Site{
			BrandColor: "#4f46e5",
			Layout:     config.Layout{RankDir: "TB", NodeSep: 50, RankSep: 80},
		},
		Roadmaps: []config.Roadmap{
			{
				ID: "test", Title: "Test Roadmap",
				Nodes: []*config.Node{{ID: "a", Title: "Node A", Type: config.NodeTypeRequired}},
			},
		},
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

func TestRenderRoadmapPage_progressSyncEnabled(t *testing.T) {
	cfg, g, lr := buildMinimalPageFixture(t)
	cfg.Site.ProgressSync = config.ProgressSync{
		Enabled:  true,
		Endpoint: "https://api.example.com/sync",
	}
	html, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nil, nil, "/", "../", false)
	if err != nil {
		t.Fatalf("RenderRoadmapPage: %v", err)
	}
	if !strings.Contains(html, `"enabled":true`) {
		t.Errorf("expected enabled:true in SITE_CONFIG")
	}
	if !strings.Contains(html, `"endpoint":"https://api.example.com/sync"`) {
		t.Errorf("expected endpoint in SITE_CONFIG")
	}
}

func TestRenderRoadmapPage_progressSyncDisabled(t *testing.T) {
	cfg, g, lr := buildMinimalPageFixture(t)
	html, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nil, nil, "/", "../", false)
	if err != nil {
		t.Fatalf("RenderRoadmapPage: %v", err)
	}
	if !strings.Contains(html, `"enabled":false`) {
		t.Errorf("expected enabled:false in SITE_CONFIG")
	}
}

func TestRenderRoadmapPage_progressSyncEndpointTrimsTrailingSlash(t *testing.T) {
	cfg, g, lr := buildMinimalPageFixture(t)
	cfg.Site.ProgressSync = config.ProgressSync{
		Enabled:  true,
		Endpoint: "https://api.example.com/sync/",
	}
	html, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nil, nil, "/", "../", false)
	if err != nil {
		t.Fatalf("RenderRoadmapPage: %v", err)
	}
	if strings.Contains(html, `"endpoint":"https://api.example.com/sync/"`) {
		t.Errorf("trailing slash should be trimmed from endpoint")
	}
	if !strings.Contains(html, `"endpoint":"https://api.example.com/sync"`) {
		t.Errorf("trimmed endpoint should be present in SITE_CONFIG")
	}
}

func TestRenderIndexPage_progressSyncEnabled(t *testing.T) {
	cfg, g, _ := buildMinimalPageFixture(t)
	cfg.Site.ProgressSync = config.ProgressSync{
		Enabled:  true,
		Endpoint: "https://api.example.com/sync",
	}
	html, err := render.RenderIndexPage(web.FS, cfg, "/", map[string]*graph.Graph{"test": g})
	if err != nil {
		t.Fatalf("RenderIndexPage: %v", err)
	}
	if !strings.Contains(html, `"enabled":true`) {
		t.Errorf("expected enabled:true in index SITE_CONFIG")
	}
	if !strings.Contains(html, `"endpoint":"https://api.example.com/sync"`) {
		t.Errorf("expected endpoint in index SITE_CONFIG")
	}
}

func TestRenderIndexPage_progressSyncDisabled(t *testing.T) {
	cfg, g, _ := buildMinimalPageFixture(t)
	html, err := render.RenderIndexPage(web.FS, cfg, "/", map[string]*graph.Graph{"test": g})
	if err != nil {
		t.Fatalf("RenderIndexPage: %v", err)
	}
	if !strings.Contains(html, `"enabled":false`) {
		t.Errorf("expected enabled:false in index SITE_CONFIG")
	}
}
