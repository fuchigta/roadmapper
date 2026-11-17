package command

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
)

func TestFindUnresolved(t *testing.T) {
	cfg := &config.Config{Roadmaps: []config.Roadmap{
		{ID: "r1", Nodes: []*config.Node{
			{ID: "ok", Title: "OK", Type: config.NodeTypeRequired},
			{ID: "missing", Title: "M", Type: config.NodeTypeRequired},
			{ID: "draft", Title: "D", Type: config.NodeTypeRequired, Draft: true},
		}},
		{ID: "r2", Nodes: []*config.Node{
			{ID: "x", Title: "X", Type: config.NodeTypeRequired, Content: "sub/none"},
			{ID: "y", Title: "Y", Type: config.NodeTypeRequired, Content: "sub/page"},
		}},
	}}
	graphs, err := buildGraphs(cfg)
	if err != nil {
		t.Fatal(err)
	}
	docs := map[string]*content.Doc{"ok": {}, "sub/page": {}}
	tests := []struct {
		name      string
		draftSets []map[string]bool
		want      []unresolvedEntry
	}{
		{"下書きは除外", []map[string]bool{{"draft": true}, nil}, []unresolvedEntry{
			{roadmapID: "r1", nodeID: "missing"},
			{roadmapID: "r2", nodeID: "x", content: "sub/none"},
		}},
		{"下書き指定なしなら含める", []map[string]bool{nil, nil}, []unresolvedEntry{
			{roadmapID: "r1", nodeID: "missing"},
			{roadmapID: "r1", nodeID: "draft"},
			{roadmapID: "r2", nodeID: "x", content: "sub/none"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findUnresolved(cfg, graphs, docs, tt.draftSets)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestCheckRoadmapChangelog(t *testing.T) {
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	nodes := []*config.Node{
		{ID: "a", Title: "A", Type: config.NodeTypeRequired},
		{ID: "d", Title: "D", Type: config.NodeTypeRequired},
	}
	fm := func(updated string) *content.Doc {
		return &content.Doc{Frontmatter: content.Frontmatter{Updated: updated}}
	}
	tests := []struct {
		name      string
		changelog []config.ChangelogEntry
		docs      map[string]*content.Doc
		drafts    map[string]bool
		want      int
		wantErrIn string
	}{
		{"警告なし", []config.ChangelogEntry{{Date: "2026-06-01", Summary: "s", Nodes: []string{"a"}}},
			map[string]*content.Doc{"a": fm("2026-06-10")}, nil, 0, ""},
		{"未来日付と古い updated", []config.ChangelogEntry{{Date: "2026-07-01", Summary: "s", Nodes: []string{"a"}}},
			map[string]*content.Doc{"a": fm("2026-06-10")}, nil, 2, ""},
		{"下書きだけを参照する項目", []config.ChangelogEntry{{Date: "2026-06-01", Summary: "s", Nodes: []string{"d"}}},
			map[string]*content.Doc{"d": fm("2020-01-01")}, map[string]bool{"d": true}, 1, ""},
		{"下書きの不正な frontmatter は検査しない", nil,
			map[string]*content.Doc{"d": fm("bad")}, map[string]bool{"d": true}, 0, ""},
		{"frontmatter の日付不正はエラー", nil,
			map[string]*content.Doc{"a": fm("bad")}, nil, 0, `ロードマップ "r" の改版履歴が不正`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := &config.Roadmap{ID: "r", Nodes: nodes, Changelog: tt.changelog}
			g, err := graph.Build(rm)
			if err != nil {
				t.Fatal(err)
			}
			got, err := checkRoadmapChangelog(rm, g, tt.docs, tt.drafts, now)
			if tt.wantErrIn != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrIn) {
					t.Fatalf("err = %v, want containing %q", err, tt.wantErrIn)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("warnings = %d, want %d", got, tt.want)
			}
		})
	}
}
