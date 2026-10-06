package render_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/render"
	"github.com/fuchigta/roadmapper/web"
)

func boolp(b bool) *bool { return &b }

func TestRenderAnalyticsHead_golden(t *testing.T) {
	tests := []struct {
		name string
		a    config.Analytics
	}{
		{"umami", config.Analytics{Provider: "umami", ScriptURL: "https://u.example.com/script.js", SiteID: "abc-123", Domains: []string{"a.example.com", "b.example.com"}, ExcludeSearch: boolp(true)}},
		{"plausible", config.Analytics{Provider: "plausible", ScriptURL: "https://plausible.io/js/script.js", SiteID: "example.com"}},
		{"goatcounter", config.Analytics{Provider: "goatcounter", ScriptURL: "https://gc.zgo.at/count.js", SiteID: "https://x.goatcounter.com/count"}},
		{"custom", config.Analytics{Provider: "custom", Head: `<script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXX"></script>`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(render.RenderAnalyticsHead(tt.a))
			path := filepath.Join("testdata", "analytics_"+tt.name+".html")
			if os.Getenv("UPDATE_GOLDEN") == "1" {
				if err := os.MkdirAll("testdata", 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("golden 読み込み失敗 (UPDATE_GOLDEN=1 で生成): %v", err)
			}
			if got != string(want) {
				t.Errorf("output mismatch:\ngot:  %s\nwant: %s", got, want)
			}
		})
	}
}

func TestRenderAnalyticsHead_disabled(t *testing.T) {
	if got := render.RenderAnalyticsHead(config.Analytics{}); got != "" {
		t.Errorf("無効時は空のはず: %q", got)
	}
}

func TestRenderAnalyticsHead_escapesAttributes(t *testing.T) {
	got := string(render.RenderAnalyticsHead(config.Analytics{
		Provider:  "umami",
		ScriptURL: `https://x.example.com/s.js"><script>alert(1)</script>`,
		SiteID:    `a"b<c>`,
		Domains:   []string{`d"e`},
	}))
	for _, bad := range []string{`"><script>alert`, `a"b<c>`, `d"e`} {
		if strings.Contains(got, bad) {
			t.Errorf("エスケープされていない %q in %s", bad, got)
		}
	}
}

func TestRenderAnalyticsHead_events(t *testing.T) {
	for _, p := range []string{"umami", "plausible", "goatcounter"} {
		t.Run(p, func(t *testing.T) {
			a := config.Analytics{Provider: p, ScriptURL: "https://x.example.com/s.js", SiteID: "id"}
			if !strings.Contains(string(render.RenderAnalyticsHead(a)), "roadmapperTrack") {
				t.Error("events 未指定 (既定 true) ではアダプタが出るはず")
			}
			a.Events = boolp(false)
			got := string(render.RenderAnalyticsHead(a))
			if strings.Contains(got, "roadmapperTrack") {
				t.Error("events: false ではアダプタが出ないはず")
			}
			if !strings.Contains(got, "https://x.example.com/s.js") {
				t.Error("events: false でもスクリプトタグは出るはず")
			}
		})
	}
}

func TestRenderPages_analyticsHead(t *testing.T) {
	enabled := config.Analytics{Provider: "plausible", ScriptURL: "https://plausible.io/js/script.js", SiteID: "example.com"}
	tests := []struct {
		name       string
		a          config.Analytics
		wantTag    bool
		wantEvents string
	}{
		{"有効", enabled, true, "analyticsEvents: true"},
		{"events false", func() config.Analytics { a := enabled; a.Events = boolp(false); return a }(), true, "analyticsEvents: false"},
		{"無効", config.Analytics{}, false, "analyticsEvents: false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, g, lr := buildMinimalPageFixture(t)
			cfg.Site.Analytics = tt.a

			page, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nil, nil, "/", "../", false, nil, nil, nil)
			if err != nil {
				t.Fatalf("RenderRoadmapPage: %v", err)
			}
			index, err := render.RenderIndexPage(web.FS, cfg, "/", map[string]*graph.Graph{"test": g}, nil)
			if err != nil {
				t.Fatalf("RenderIndexPage: %v", err)
			}
			for name, html := range map[string]string{"roadmap": page, "index": index} {
				head, _, ok := strings.Cut(html, "</head>")
				if !ok {
					t.Fatalf("%s: </head> がない", name)
				}
				if has := strings.Contains(head, `data-domain="example.com"`); has != tt.wantTag {
					t.Errorf("%s: head のタグ有無 = %v, want %v", name, has, tt.wantTag)
				}
				if !strings.Contains(strings.Join(strings.Fields(html), " "), tt.wantEvents) {
					t.Errorf("%s: %q が SITE_CONFIG にない", name, tt.wantEvents)
				}
			}
		})
	}
}
