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

	ws := checkEntries(entries, nodeDocs, today)
	for _, id := range sortedNodeIDs(nodeDocs) {
		ws = append(ws, checkNodeFrontmatter(id, nodeDocs[id].Frontmatter, today)...)
	}

	sort.SliceStable(ws, func(i, j int) bool {
		if ws[i].NodeID != ws[j].NodeID {
			return ws[i].NodeID < ws[j].NodeID
		}
		return ws[i].Message < ws[j].Message
	})
	return ws
}

// checkEntries は roadmap.yml の項目の未来日付と、関連ノードの updated が項目より古いことを検査する。
func checkEntries(entries []config.ChangelogEntry, nodeDocs map[string]*content.Doc, today time.Time) []Warning {
	var ws []Warning
	for i, e := range entries {
		ed, err := parseDate(e.Date)
		if err != nil {
			continue
		}
		if ed.After(today) {
			ws = append(ws, Warning{Message: fmt.Sprintf("changelog[%d] の日付 %s が未来です", i, e.Date)})
		}
		ws = append(ws, checkEntryNodes(i, e, ed, nodeDocs)...)
	}
	return ws
}

// checkEntryNodes は roadmap.yml の項目 entries[i] (日付 ed) より updated が古い関連ノードを警告する。
func checkEntryNodes(i int, e config.ChangelogEntry, ed time.Time, nodeDocs map[string]*content.Doc) []Warning {
	var ws []Warning
	for _, id := range dedupe(e.Nodes) {
		ud, ok := docUpdated(nodeDocs[id])
		if ok && ud.Before(ed) {
			ws = append(ws, Warning{NodeID: id, Message: fmt.Sprintf(
				"updated %s が roadmap.yml の changelog[%d] (%s) より古いです", nodeDocs[id].Frontmatter.Updated, i, e.Date)})
		}
	}
	return ws
}

// docUpdated は doc の frontmatter updated を返す。doc がない・updated が空・パース不可なら false。
func docUpdated(doc *content.Doc) (time.Time, bool) {
	if doc == nil || doc.Frontmatter.Updated == "" {
		return time.Time{}, false
	}
	ud, err := parseDate(doc.Frontmatter.Updated)
	if err != nil {
		return time.Time{}, false
	}
	return ud, true
}

// checkNodeFrontmatter はノード frontmatter の changes / updated の未来日付と、
// updated が changes の最新日より古いことを検査する。
func checkNodeFrontmatter(id string, fm content.Frontmatter, today time.Time) []Warning {
	ws, latestChange, latestChangeStr := checkChanges(id, fm.Changes, today)
	if fm.Updated == "" {
		return ws
	}
	ud, err := parseDate(fm.Updated)
	if err != nil {
		return ws
	}
	if ud.After(today) {
		ws = append(ws, Warning{NodeID: id, Message: fmt.Sprintf("updated %s が未来です", fm.Updated)})
	}
	if !latestChange.IsZero() && ud.Before(latestChange) {
		ws = append(ws, Warning{NodeID: id, Message: fmt.Sprintf("updated %s が changes の最新日 %s より古いです", fm.Updated, latestChangeStr)})
	}
	return ws
}

// checkChanges は changes の未来日付を警告し、パースできた changes の最新日 (とその元の文字列) を返す。
func checkChanges(id string, changes []content.Change, today time.Time) (ws []Warning, latest time.Time, latestStr string) {
	for ci, c := range changes {
		cd, err := parseDate(c.Date)
		if err != nil {
			continue
		}
		if cd.After(today) {
			ws = append(ws, Warning{NodeID: id, Message: fmt.Sprintf("changes[%d] の日付 %s が未来です", ci, c.Date)})
		}
		if cd.After(latest) {
			latest, latestStr = cd, c.Date
		}
	}
	return ws, latest, latestStr
}

// Recent は updated のうち now から days 日以内 (未来日付は除く) のノード ID の集合を返す。
func Recent(updated map[string]time.Time, now time.Time, days int) map[string]bool {
	out := map[string]bool{}
	limit := time.Duration(days) * 24 * time.Hour
	for id, d := range updated {
		if d.After(now) {
			continue
		}
		if now.Sub(d) <= limit {
			out[id] = true
		}
	}
	return out
}

// StripNodes は entries の関連ノード ID から exclude に含まれるものを外した写しを返す。
// 項目自体は残す。下書きノードへの参照を公開用の履歴から取り除くのに使う。
// exclude が空なら entries をそのまま返す。
func StripNodes(entries []config.ChangelogEntry, exclude map[string]bool) []config.ChangelogEntry {
	if len(exclude) == 0 {
		return entries
	}
	out := make([]config.ChangelogEntry, len(entries))
	for i, e := range entries {
		out[i] = e
		if len(e.Nodes) == 0 {
			continue
		}
		kept := make([]string, 0, len(e.Nodes))
		for _, id := range e.Nodes {
			if !exclude[id] {
				kept = append(kept, id)
			}
		}
		out[i].Nodes = kept
	}
	return out
}

// CheckDrafts は nodes が下書きノードだけを参照している項目の警告を返す。
// そうした項目は公開ビルドで関連ノードを持たない履歴になる。
func CheckDrafts(entries []config.ChangelogEntry, drafts map[string]bool) []Warning {
	var ws []Warning
	for i, e := range entries {
		if len(e.Nodes) == 0 {
			continue
		}
		onlyDraft := true
		for _, id := range e.Nodes {
			if !drafts[id] {
				onlyDraft = false
				break
			}
		}
		if onlyDraft {
			ws = append(ws, Warning{Message: fmt.Sprintf(
				"changelog[%d] (%s) の nodes が下書きノードだけを参照しています (公開ビルドでは参照が外れます)", i, e.Date)})
		}
	}
	return ws
}
