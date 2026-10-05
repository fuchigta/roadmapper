package meta

import (
	"strings"
	"testing"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
)

func testCfg(t *testing.T, siteURL string) (*config.Config, map[string]*graph.Graph) {
	t.Helper()
	cfg := &config.Config{
		Site: config.Site{Title: "T", Description: "D", SiteURL: siteURL, BasePath: "/repo"},
		Roadmaps: []config.Roadmap{{
			ID: "fe", Title: "Frontend",
			Nodes: []*config.Node{{ID: "html", Title: "HTML"}, {ID: "css", Title: "CSS"}},
		}},
	}
	g, err := graph.Build(&cfg.Roadmaps[0])
	if err != nil {
		t.Fatal(err)
	}
	return cfg, map[string]*graph.Graph{"fe": g}
}

func day(s string) time.Time {
	d, _ := time.Parse(config.ChangelogDateLayout, s)
	return d
}

func TestRenderRSS_emptySiteURL(t *testing.T) {
	cfg, graphs := testCfg(t, "")
	got, err := RenderRSS(cfg, graphs, nil)
	if err != nil || got != "" {
		t.Fatalf("got %q, %v; want empty", got, err)
	}
}

func TestRenderRSS_nodeItemsWithoutChangelog(t *testing.T) {
	cfg, graphs := testCfg(t, "https://example.com")
	got, err := RenderRSS(cfg, graphs, map[string]*changelog.Log{"fe": {}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"<guid>https://example.com/repo/fe/index.html#html</guid>",
		"<guid>https://example.com/repo/fe/index.html#css</guid>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRenderRSS_changelogItems(t *testing.T) {
	cfg, graphs := testCfg(t, "https://example.com")
	log := &changelog.Log{
		Items: []changelog.Item{
			{Date: day("2026-09-20"), Summary: "フォーム追記", NodeIDs: []string{"html"}, FromNode: "html"},
			{Date: day("2026-09-01"), Summary: "全体見直し", NodeIDs: []string{"html", "css"}},
			{Date: day("2026-08-01"), Summary: "CSS 追加", NodeIDs: []string{"css"}},
		},
		Latest: day("2026-09-20"),
	}
	got, err := RenderRSS(cfg, graphs, map[string]*changelog.Log{"fe": log})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"<title>HTML: フォーム追記</title>",
		"<link>https://example.com/repo/fe/index.html#html</link>",
		"<title>全体見直し</title>",
		"<link>https://example.com/repo/fe/index.html</link>",
		"<link>https://example.com/repo/fe/index.html#css</link>",
		"<pubDate>Sun, 20 Sep 2026 00:00:00 +0000</pubDate>",
		"<lastBuildDate>Sun, 20 Sep 2026 00:00:00 +0000</lastBuildDate>",
		"#changelog-2026-09-20-0</guid>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "<guid>https://example.com/repo/fe/index.html#html</guid>") {
		t.Errorf("node items should not be emitted when changelog exists")
	}
	if n := strings.Count(got, "<item>"); n != 3 {
		t.Errorf("item count = %d, want 3", n)
	}
}
