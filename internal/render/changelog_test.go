package render_test

import (
	"strings"
	"testing"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/render"
	"github.com/fuchigta/roadmapper/web"
)

func sampleLog() *changelog.Log {
	d1 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	items := []changelog.Item{
		{Date: d1, Summary: "Node A を改訂 <script>alert(1)</script>", NodeIDs: []string{"a"}, FromNode: "a"},
		{Date: d2, Summary: "全体を見直し", NodeIDs: []string{"a"}},
	}
	return &changelog.Log{
		Items:       items,
		NodeItems:   map[string][]changelog.Item{"a": items},
		NodeUpdated: map[string]time.Time{"a": d1},
		Latest:      d1,
	}
}

func TestRenderRoadmapPage_changelog(t *testing.T) {
	cfg, g, lr := buildMinimalPageFixture(t)
	nodeHTML := map[string]string{"a": "<p>body</p>"}

	t.Run("履歴なし", func(t *testing.T) {
		for _, lg := range []*changelog.Log{nil, {}} {
			html, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nodeHTML, nil, "/", "../", false, lg)
			if err != nil {
				t.Fatal(err)
			}
			for _, s := range []string{"changelog-btn", "changelog-tpl", "changelog-latest", "node-changelog"} {
				if strings.Contains(html, s) {
					t.Errorf("履歴なしなのに %q が出力されている", s)
				}
			}
		}
	})

	t.Run("履歴あり", func(t *testing.T) {
		html, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nodeHTML, nil, "/", "../", false, sampleLog())
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range []string{
			`id="changelog-btn"`,
			`2026-05-01</time> 更新: Node A: Node A を改訂`, // FromNode はノードタイトルを前置
			`<template id="changelog-tpl">`,
			`全体を見直し`,
			`data-node="a"`,
			`node-changelog`, // ノード別履歴 (ROADMAP_DATA の html 内)
			`2026-05-01`,
		} {
			if !strings.Contains(html, s) {
				t.Errorf("出力に %q が含まれない", s)
			}
		}
		if strings.Contains(html, "<script>alert(1)</script>") {
			t.Errorf("summary の <script> がエスケープされていない")
		}
		if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
			t.Errorf("パネル用履歴にエスケープ済みの summary が見つからない")
		}
		if strings.Contains(html, "ZgotmplZ") {
			t.Errorf("unexpected ZgotmplZ")
		}
	})

	t.Run("呼び出し元の nodeHTML を変更しない", func(t *testing.T) {
		if _, err := render.RenderRoadmapPage(web.FS, cfg, &cfg.Roadmaps[0], g, lr, nodeHTML, nil, "/", "../", false, sampleLog()); err != nil {
			t.Fatal(err)
		}
		if nodeHTML["a"] != "<p>body</p>" {
			t.Errorf("nodeHTML が変更された: %q", nodeHTML["a"])
		}
	})
}

func TestRenderIndexPage_latestUpdated(t *testing.T) {
	cfg, g, _ := buildMinimalPageFixture(t)
	graphs := map[string]*graph.Graph{"test": g}

	html, err := render.RenderIndexPage(web.FS, cfg, "/", graphs, map[string]time.Time{"test": time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html, "最終更新: <time>2026-05-01</time>") {
		t.Errorf("最終更新日が表示されていない")
	}

	for _, latest := range []map[string]time.Time{nil, {"test": {}}} {
		html, err = render.RenderIndexPage(web.FS, cfg, "/", graphs, latest)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(html, "最終更新") {
			t.Errorf("zero の Latest なのに最終更新が表示されている")
		}
	}
}
