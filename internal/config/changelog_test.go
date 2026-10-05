package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
)

func TestValidate_changelog(t *testing.T) {
	mk := func(entries ...config.ChangelogEntry) *config.Config {
		return &config.Config{
			Site: config.Site{Title: "T"},
			Roadmaps: []config.Roadmap{{
				ID: "fe", Title: "FE",
				Nodes:     []*config.Node{{ID: "html", Title: "HTML", Type: config.NodeTypeRequired}},
				Changelog: entries,
			}},
		}
	}
	tests := []struct {
		name      string
		cfg       *config.Config
		wantField string // 空なら成功を期待
	}{
		{"正常", mk(config.ChangelogEntry{Date: "2026-10-05", Summary: "s", Nodes: []string{"html"}}), ""},
		{"date 空", mk(config.ChangelogEntry{Summary: "s"}), "roadmaps[fe].changelog[0].date"},
		{"date 不正", mk(config.ChangelogEntry{Date: "2026-13-01", Summary: "s"}), "roadmaps[fe].changelog[0].date"},
		{"summary 空", mk(config.ChangelogEntry{Date: "2026-10-05"}), "roadmaps[fe].changelog[0].summary"},
		{"未知ノード", mk(config.ChangelogEntry{Date: "2026-10-05", Summary: "s", Nodes: []string{"html", "nope"}}), "roadmaps[fe].changelog[0].nodes[1]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.Validate(tt.cfg)
			if tt.wantField == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantField) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantField)
			}
		})
	}
}

func TestLoad_changelogUnquotedDate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "roadmap.yml")
	yml := `site:
  title: T
roadmaps:
  - id: fe
    title: FE
    nodes:
      - {id: html, title: HTML, type: required}
    changelog:
      - date: 2026-10-05
        summary: 追加
        nodes: [html]
`
	if err := os.WriteFile(path, []byte(yml), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	got := cfg.Roadmaps[0].Changelog
	if len(got) != 1 || got[0].Date != "2026-10-05" || got[0].Summary != "追加" || got[0].Nodes[0] != "html" {
		t.Errorf("changelog = %+v", got)
	}
}
