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
		"edge":    d("2026-09-05"),
		"old":     d("2026-09-04"),
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

// TestRecentBoundary は「days 日以内」を暦日の差で数えることを固定する。now の日付 (now 自身の
// ロケーション) から数えてちょうど days 日前の日付は含み、days+1 日前は含まない。時刻は結果に
// 影響しない。updated は parseDate と同じく日付文字列を UTC 0 時としてパースした値を使う。
func TestRecentBoundary(t *testing.T) {
	jst := time.FixedZone("JST", 9*60*60)
	est := time.FixedZone("EST", -5*60*60)
	date := func(s string) time.Time {
		v, err := parseDate(s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	// now の日付はいずれも 2026-10-05 (now のロケーションで)。
	nows := []struct {
		name string
		now  time.Time
	}{
		{"UTC 0時", time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)},
		{"UTC 正午", time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)},
		{"UTC 23:59:59.999999999", time.Date(2026, 10, 5, 23, 59, 59, 999999999, time.UTC)},
		// UTC では 2026-10-04 20:00。旧仕様 (経過時間) では当日の updated が未来扱いで外れていた
		{"JST 早朝", time.Date(2026, 10, 5, 5, 0, 0, 0, jst)},
		{"JST 深夜", time.Date(2026, 10, 5, 23, 30, 0, 0, jst)},
		// UTC では 2026-10-06 01:00。日付は UTC ではなく now のロケーションで読む
		{"EST 夜", time.Date(2026, 10, 5, 20, 0, 0, 0, est)},
	}
	cases := []struct {
		name    string
		updated string
		want    bool
	}{
		{"同日は含む", "2026-10-05", true},
		{"ちょうど 30 日前は含む", "2026-09-05", true},
		{"29 日前は含む", "2026-09-06", true},
		{"31 日前は含まない", "2026-09-04", false},
		{"翌日 (未来) は含まない", "2026-10-06", false},
	}
	for _, n := range nows {
		for _, c := range cases {
			t.Run(n.name+"/"+c.name, func(t *testing.T) {
				got := Recent(map[string]time.Time{"n": date(c.updated)}, n.now, 30)
				if got["n"] != c.want {
					t.Errorf("Recent(%s, now=%v, 30)[n] = %v, want %v", c.updated, n.now, got["n"], c.want)
				}
			})
		}
	}
}

// TestRecentCalendarVsElapsed は旧仕様 (経過時間 days*24h) との違いを示す。日付差は 30 日だが
// 経過時間では 30*24h を超える (now がその日の遅い時刻) 場合も含む。
func TestRecentCalendarVsElapsed(t *testing.T) {
	now := time.Date(2026, 10, 5, 23, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	if elapsed := now.Sub(updated); elapsed <= 30*24*time.Hour {
		t.Fatalf("前提: 経過時間 %v が 30*24h を超えていること", elapsed)
	}
	if !Recent(map[string]time.Time{"n": updated}, now, 30)["n"] {
		t.Error("日付差がちょうど 30 日なら経過時間によらず含むこと")
	}
}

func TestRecentZeroDays(t *testing.T) {
	now := time.Date(2026, 10, 5, 18, 0, 0, 0, time.UTC)
	got := Recent(map[string]time.Time{
		"today":     time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		"yesterday": time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC),
	}, now, 0)
	if want := map[string]bool{"today": true}; !reflect.DeepEqual(got, want) {
		t.Errorf("days=0 は当日のみ: got %v, want %v", got, want)
	}
}

func TestStripNodes(t *testing.T) {
	entries := []config.ChangelogEntry{
		{Date: "2026-01-01", Summary: "a", Nodes: []string{"x", "d"}},
		{Date: "2026-01-02", Summary: "b", Nodes: []string{"d"}},
		{Date: "2026-01-03", Summary: "c"},
	}
	tests := []struct {
		name    string
		exclude map[string]bool
		want    [][]string
	}{
		{"除外なし", nil, [][]string{{"x", "d"}, {"d"}, nil}},
		{"d を除外", map[string]bool{"d": true}, [][]string{{"x"}, {}, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripNodes(entries, tt.exclude)
			if len(got) != len(entries) {
				t.Fatalf("len = %d", len(got))
			}
			for i, w := range tt.want {
				if len(got[i].Nodes) != len(w) {
					t.Errorf("entry %d: nodes = %v, want %v", i, got[i].Nodes, w)
				}
				if got[i].Summary != entries[i].Summary {
					t.Errorf("entry %d: summary が変わっています", i)
				}
			}
		})
	}
	if len(entries[0].Nodes) != 2 {
		t.Error("入力が書き換えられています")
	}
}

func TestCheckDrafts(t *testing.T) {
	entries := []config.ChangelogEntry{
		{Date: "2026-01-01", Summary: "mixed", Nodes: []string{"x", "d"}},
		{Date: "2026-01-02", Summary: "draft only", Nodes: []string{"d"}},
		{Date: "2026-01-03", Summary: "no nodes"},
	}
	got := CheckDrafts(entries, map[string]bool{"d": true})
	if len(got) != 1 || !strings.Contains(got[0].Message, "changelog[1]") {
		t.Fatalf("got %+v, want changelog[1] のみ", got)
	}
	if got := CheckDrafts(entries, nil); len(got) != 0 {
		t.Errorf("下書きなしなら警告なし: %+v", got)
	}
}

// 下書きの frontmatter (updated / changes) は呼び出し側が入力から除外すれば集約されない。
func TestBuild_draftDocsExcluded(t *testing.T) {
	pub := &content.Doc{Frontmatter: content.Frontmatter{Updated: "2026-02-01"}}
	draft := &content.Doc{Frontmatter: content.Frontmatter{
		Draft: true, Updated: "2026-03-01",
		Changes: []content.Change{{Date: "2026-03-01", Summary: "SECRET"}},
	}}
	docs := map[string]*content.Doc{"pub": pub, "draft": draft}
	delete(docs, "draft")
	log, err := Build(StripNodes(
		[]config.ChangelogEntry{{Date: "2026-01-01", Summary: "s", Nodes: []string{"draft"}}},
		map[string]bool{"draft": true}), docs)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := log.NodeUpdated["draft"]; ok {
		t.Error("下書きノードが NodeUpdated に含まれています")
	}
	if _, ok := log.NodeItems["draft"]; ok {
		t.Error("下書きノードが NodeItems に含まれています")
	}
	for _, it := range log.Items {
		if it.Summary == "SECRET" {
			t.Error("下書きの changes が集約されています")
		}
	}
	if len(log.Items) != 1 {
		t.Errorf("roadmap.yml 由来の項目は残るはず: %d", len(log.Items))
	}
}

func TestCheckNodeFrontmatter(t *testing.T) {
	today := day("2026-06-15")
	change := func(date string) content.Change { return content.Change{Date: date, Summary: "c"} }
	tests := []struct {
		name string
		fm   content.Frontmatter
		want []Warning
	}{
		{"空", content.Frontmatter{}, nil},
		{"updated のみ (過去)", content.Frontmatter{Updated: "2026-06-01"}, nil},
		{"updated が当日は未来扱いしない", content.Frontmatter{Updated: "2026-06-15"}, nil},
		{"updated 不正でも changes の未来日付は警告", content.Frontmatter{Updated: "bad", Changes: []content.Change{change("2026-07-01")}},
			[]Warning{{"a", "changes[0] の日付 2026-07-01 が未来です"}}},
		{"changes がすべて不正なら最新日比較をしない", content.Frontmatter{Updated: "2026-01-01", Changes: []content.Change{change("bad")}}, nil},
		{"最新の changes と比較する", content.Frontmatter{Updated: "2026-03-01", Changes: []content.Change{change("2026-02-01"), change("2026-04-01"), change("2026-01-01")}},
			[]Warning{{"a", "updated 2026-03-01 が changes の最新日 2026-04-01 より古いです"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkNodeFrontmatter("a", tt.fm, today)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v\nwant %+v", got, tt.want)
			}
		})
	}
}
