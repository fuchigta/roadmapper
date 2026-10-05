// Package changelog は roadmap.yml と content frontmatter の改版履歴を集約・検証する。
// 純粋関数のみで構成し、ファイル I/O は行わない。
package changelog

import (
	"fmt"
	"sort"
	"time"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
)

// Item は改版履歴の 1 項目。
type Item struct {
	Date    time.Time
	Summary string
	NodeIDs []string
	// FromNode はノード frontmatter 由来ならそのノード ID、roadmap.yml 由来なら空。
	FromNode string
}

// Log は集約済みの改版履歴。
type Log struct {
	// Items は日付降順。同日は roadmap.yml 由来が先、その後は入力順。
	Items []Item
	// NodeItems はノードごとの関連履歴 (日付降順)。
	NodeItems map[string][]Item
	// NodeUpdated はノードごとの最新日。
	NodeUpdated map[string]time.Time
	// Latest は全体の最新日 (履歴がなければ zero)。
	Latest time.Time
}

// Warning は検証警告 1 件。NodeID が空なら roadmap.yml 項目に対する警告。
type Warning struct {
	NodeID  string
	Message string
}

func parseDate(s string) (time.Time, error) {
	return time.Parse(config.ChangelogDateLayout, s)
}

func sortedNodeIDs(nodeDocs map[string]*content.Doc) []string {
	ids := make([]string, 0, len(nodeDocs))
	for id, d := range nodeDocs {
		if d != nil {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids
}

func dedupe(ids []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func sortDesc(items []Item) {
	sort.SliceStable(items, func(i, j int) bool { return items[i].Date.After(items[j].Date) })
}

// Build は roadmap.yml の項目とノード frontmatter から Log を組み立てる。
// nodeDocs はノード ID → 解決済み Doc (未解決ノードは含めない)。
func Build(entries []config.ChangelogEntry, nodeDocs map[string]*content.Doc) (*Log, error) {
	log := &Log{
		NodeItems:   map[string][]Item{},
		NodeUpdated: map[string]time.Time{},
	}
	bump := func(id string, d time.Time) {
		if cur, ok := log.NodeUpdated[id]; !ok || d.After(cur) {
			log.NodeUpdated[id] = d
		}
	}

	for i, e := range entries {
		d, err := parseDate(e.Date)
		if err != nil {
			return nil, fmt.Errorf("changelog[%d] の日付 %q が不正: %w", i, e.Date, err)
		}
		it := Item{Date: d, Summary: e.Summary, NodeIDs: dedupe(e.Nodes)}
		log.Items = append(log.Items, it)
		for _, id := range it.NodeIDs {
			log.NodeItems[id] = append(log.NodeItems[id], it)
			bump(id, d)
		}
	}

	for _, id := range sortedNodeIDs(nodeDocs) {
		fm := nodeDocs[id].Frontmatter
		if fm.Updated != "" {
			d, err := parseDate(fm.Updated)
			if err != nil {
				return nil, fmt.Errorf("ノード %q の frontmatter updated %q が不正: %w", id, fm.Updated, err)
			}
			bump(id, d)
		}
		for ci, c := range fm.Changes {
			d, err := parseDate(c.Date)
			if err != nil {
				return nil, fmt.Errorf("ノード %q の frontmatter changes[%d] の日付 %q が不正: %w", id, ci, c.Date, err)
			}
			it := Item{Date: d, Summary: c.Summary, NodeIDs: []string{id}, FromNode: id}
			log.Items = append(log.Items, it)
			log.NodeItems[id] = append(log.NodeItems[id], it)
			bump(id, d)
		}
	}

	sortDesc(log.Items)
	for _, items := range log.NodeItems {
		sortDesc(items)
	}
	if len(log.Items) > 0 {
		log.Latest = log.Items[0].Date
	}
	for _, d := range log.NodeUpdated {
		if d.After(log.Latest) {
			log.Latest = d
		}
	}
	return log, nil
}

// Check は改版履歴の整合性に関する警告を返す。日付パース不可の項目は無視する。
// 結果はノード ID、メッセージ順で決定的にソートされる。
func Check(entries []config.ChangelogEntry, nodeDocs map[string]*content.Doc, now time.Time) []Warning {
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	var ws []Warning
	add := func(id, msg string) { ws = append(ws, Warning{NodeID: id, Message: msg}) }

	for i, e := range entries {
		ed, err := parseDate(e.Date)
		if err != nil {
			continue
		}
		if ed.After(today) {
			add("", fmt.Sprintf("changelog[%d] の日付 %s が未来です", i, e.Date))
		}
		for _, id := range dedupe(e.Nodes) {
			doc := nodeDocs[id]
			if doc == nil || doc.Frontmatter.Updated == "" {
				continue
			}
			ud, err := parseDate(doc.Frontmatter.Updated)
			if err != nil {
				continue
			}
			if ud.Before(ed) {
				add(id, fmt.Sprintf("updated %s が roadmap.yml の changelog[%d] (%s) より古いです", doc.Frontmatter.Updated, i, e.Date))
			}
		}
	}

	for _, id := range sortedNodeIDs(nodeDocs) {
		fm := nodeDocs[id].Frontmatter
		var latestChange time.Time
		var latestChangeStr string
		for ci, c := range fm.Changes {
			cd, err := parseDate(c.Date)
			if err != nil {
				continue
			}
			if cd.After(today) {
				add(id, fmt.Sprintf("changes[%d] の日付 %s が未来です", ci, c.Date))
			}
			if cd.After(latestChange) {
				latestChange, latestChangeStr = cd, c.Date
			}
		}
		if fm.Updated == "" {
			continue
		}
		ud, err := parseDate(fm.Updated)
		if err != nil {
			continue
		}
		if ud.After(today) {
			add(id, fmt.Sprintf("updated %s が未来です", fm.Updated))
		}
		if !latestChange.IsZero() && ud.Before(latestChange) {
			add(id, fmt.Sprintf("updated %s が changes の最新日 %s より古いです", fm.Updated, latestChangeStr))
		}
	}

	sort.SliceStable(ws, func(i, j int) bool {
		if ws[i].NodeID != ws[j].NodeID {
			return ws[i].NodeID < ws[j].NodeID
		}
		return ws[i].Message < ws[j].Message
	})
	return ws
}
