package changelog

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
)

func day(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func doc(updated string, changes ...content.Change) *content.Doc {
	return &content.Doc{Frontmatter: content.Frontmatter{Updated: updated, Changes: changes}}
}

func summaries(items []Item) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.Summary)
	}
	return out
}

func TestBuild(t *testing.T) {
	entries := []config.ChangelogEntry{
		{Date: "2026-01-10", Summary: "r1", Nodes: []string{"html", "html"}},
		{Date: "2026-03-01", Summary: "r2"},
		{Date: "2026-02-01", Summary: "r3", Nodes: []string{"css"}},
	}
	docs := map[string]*content.Doc{
		"html": doc("2026-02-15", content.Change{Date: "2026-03-01", Summary: "n1"}, content.Change{Date: "2026-01-05", Summary: "n0"}),
		"css":  doc(""),
	}
	log, err := Build(entries, docs)
	if err != nil {
		t.Fatal(err)
	}

	// 同日 (03-01) は roadmap.yml 由来 r2 が先
	if got, want := summaries(log.Items), []string{"r2", "n1", "r3", "r1", "n0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Items = %v, want %v", got, want)
	}
	if log.Items[1].FromNode != "html" || log.Items[0].FromNode != "" {
		t.Errorf("FromNode wrong: %+v", log.Items[:2])
	}
	if !reflect.DeepEqual(log.Items[3].NodeIDs, []string{"html"}) {
		t.Errorf("NodeIDs should be deduped: %v", log.Items[3].NodeIDs)
	}
	if got, want := summaries(log.NodeItems["html"]), []string{"n1", "r1", "n0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("NodeItems[html] = %v, want %v", got, want)
	}
	if n := len(log.NodeItems["css"]); n != 1 {
		t.Errorf("NodeItems[css] len = %d", n)
	}
	if !log.NodeUpdated["html"].Equal(day("2026-03-01")) {
		t.Errorf("NodeUpdated[html] = %v", log.NodeUpdated["html"])
	}
	if !log.NodeUpdated["css"].Equal(day("2026-02-01")) {
		t.Errorf("NodeUpdated[css] = %v", log.NodeUpdated["css"])
	}
	if !log.Latest.Equal(day("2026-03-01")) {
		t.Errorf("Latest = %v", log.Latest)
	}
}

func TestBuild_updatedOnlyAndEmpty(t *testing.T) {
	log, err := Build(nil, map[string]*content.Doc{"a": doc("2026-05-01")})
	if err != nil {
		t.Fatal(err)
	}
	if len(log.Items) != 0 || !log.NodeUpdated["a"].Equal(day("2026-05-01")) || !log.Latest.Equal(day("2026-05-01")) {
		t.Errorf("unexpected: %+v", log)
	}
	log, err = Build(nil, nil)
	if err != nil || !log.Latest.IsZero() || len(log.Items) != 0 {
		t.Errorf("empty: %+v, %v", log, err)
	}
}

func TestBuild_errors(t *testing.T) {
	tests := []struct {
		name    string
		entries []config.ChangelogEntry
		docs    map[string]*content.Doc
		wantMsg string
	}{
		{"roadmap 日付不正", []config.ChangelogEntry{{Date: "2026/01/01", Summary: "x"}}, nil, "changelog[0]"},
		{"roadmap 日付空", []config.ChangelogEntry{{Date: "", Summary: "x"}}, nil, "changelog[0]"},
		{"updated 不正", nil, map[string]*content.Doc{"html": doc("bad")}, `"html"`},
		{"changes 不正", nil, map[string]*content.Doc{"css": doc("", content.Change{Date: "x", Summary: "s"})}, "changes[0]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Build(tt.entries, tt.docs)
			if err == nil || !strings.Contains(err.Error(), tt.wantMsg) {
				t.Fatalf("err = %v, want containing %q", err, tt.wantMsg)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	now := time.Date(2026, 6, 15, 23, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		entries []config.ChangelogEntry
		docs    map[string]*content.Doc
		want    []Warning
	}{
		{"警告なし", []config.ChangelogEntry{{Date: "2026-06-15", Summary: "s", Nodes: []string{"a"}}},
			map[string]*content.Doc{"a": doc("2026-06-15", content.Change{Date: "2026-06-01", Summary: "c"})}, nil},
		{"updated が changes より古い", nil,
			map[string]*content.Doc{"a": doc("2026-01-01", content.Change{Date: "2026-02-01", Summary: "c"})},
			[]Warning{{"a", "updated 2026-01-01 が changes の最新日 2026-02-01 より古いです"}}},
		{"roadmap 項目の未来日付", []config.ChangelogEntry{{Date: "2026-06-16", Summary: "s"}}, nil,
			[]Warning{{"", "changelog[0] の日付 2026-06-16 が未来です"}}},
		{"updated と changes の未来日付", nil,
			map[string]*content.Doc{"a": doc("2026-07-01", content.Change{Date: "2026-08-01", Summary: "c"})},
			[]Warning{
				{"a", "changes[0] の日付 2026-08-01 が未来です"},
				{"a", "updated 2026-07-01 が changes の最新日 2026-08-01 より古いです"},
				{"a", "updated 2026-07-01 が未来です"},
			}},
		{"updated が roadmap 項目より古い", []config.ChangelogEntry{{Date: "2026-03-01", Summary: "s", Nodes: []string{"b", "a", "none"}}},
			map[string]*content.Doc{"a": doc("2026-02-01"), "b": doc("2026-03-01"), "c": doc("")},
			[]Warning{{"a", "updated 2026-02-01 が roadmap.yml の changelog[0] (2026-03-01) より古いです"}}},
		{"updated 無しは警告しない", []config.ChangelogEntry{{Date: "2026-03-01", Summary: "s", Nodes: []string{"a"}}},
			map[string]*content.Doc{"a": doc("")}, nil},
		{"パース不可は無視", []config.ChangelogEntry{{Date: "bad", Summary: "s", Nodes: []string{"a"}}},
			map[string]*content.Doc{"a": doc("bad", content.Change{Date: "bad"})}, nil},
		{"ノード ID 順にソート", nil,
			map[string]*content.Doc{"z": doc("2027-01-01"), "a": doc("2027-01-01")},
			[]Warning{{"a", "updated 2027-01-01 が未来です"}, {"z", "updated 2027-01-01 が未来です"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Check(tt.entries, tt.docs, now)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestRecent(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	d := func(s string) time.Time {
		v, _ := time.Parse("2006-01-02", s)
		return v
	}
	updated := map[string]time.Time{
		"today":   d("2026-10-05"),
		"edge":    d("2026-09-06"),
		"old":     d("2026-09-05"),
		"future":  d("2026-10-06"),
		"ancient": d("2025-01-01"),
	}
	got := Recent(updated, now, 30)
	want := map[string]bool{"today": true, "edge": true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if len(Recent(nil, now, 30)) != 0 {
		t.Error("nil input should give empty set")
	}
}
