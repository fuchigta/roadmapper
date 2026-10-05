package render

import (
	"html/template"
	"strings"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/graph"
)

// changelogDateLayout は画面表示用の日付書式。
const changelogDateLayout = "2006-01-02"

// historyItemView はテンプレートに渡す履歴 1 項目。
type historyItemView struct {
	Date    string
	Summary string
	Nodes   []historyNodeView
}

type historyNodeView struct {
	ID    string
	Title string
}

var historyTmpl = template.Must(template.New("history").Parse(
	`{{define "items"}}<ul class="changelog-list">{{range .}}<li class="changelog-item">` +
		`<time class="changelog-date">{{.Date}}</time> <span class="changelog-summary">{{.Summary}}</span>` +
		`{{if .Nodes}} <span class="changelog-nodes">{{range .Nodes}}<a href="#{{.ID}}" data-node="{{.ID}}">{{.Title}}</a>{{end}}</span>{{end}}` +
		`</li>{{end}}</ul>{{end}}` +
		`{{define "panel"}}<section class="changelog">{{template "items" .}}</section>{{end}}` +
		`{{define "node"}}<section class="node-changelog"><h3>更新履歴</h3>` +
		`<p class="changelog-updated">最終更新: <time>{{.Updated}}</time></p>{{template "items" .Items}}</section>{{end}}`,
))

func historyItemViews(items []changelog.Item, titles map[string]string, skipNode string) []historyItemView {
	out := make([]historyItemView, 0, len(items))
	for _, it := range items {
		v := historyItemView{Date: it.Date.Format(changelogDateLayout), Summary: it.Summary}
		for _, id := range it.NodeIDs {
			if id == skipNode {
				continue
			}
			if t, ok := titles[id]; ok {
				v.Nodes = append(v.Nodes, historyNodeView{ID: id, Title: t})
			}
		}
		out = append(out, v)
	}
	return out
}

func execHistory(name string, data any) template.HTML {
	var sb strings.Builder
	if err := historyTmpl.ExecuteTemplate(&sb, name, data); err != nil {
		return ""
	}
	return template.HTML(sb.String())
}

// renderPanelHistory はサイドパネルの履歴モード用 HTML (全履歴) を返す。
func renderPanelHistory(log *changelog.Log, titles map[string]string) template.HTML {
	return execHistory("panel", historyItemViews(log.Items, titles, ""))
}

// renderNodeHistory はノード記事の末尾に付ける履歴セクション HTML を返す。履歴がなければ空。
func renderNodeHistory(log *changelog.Log, nodeID string, titles map[string]string) string {
	items := log.NodeItems[nodeID]
	if len(items) == 0 {
		return ""
	}
	return string(execHistory("node", map[string]any{
		"Updated": log.NodeUpdated[nodeID].Format(changelogDateLayout),
		"Items":   historyItemViews(items, titles, nodeID),
	}))
}

// latestLine はヘッダー表示用の最新 1 件 (日付と本文) を返す。
func latestLine(log *changelog.Log, titles map[string]string) (date, text string) {
	it := log.Items[0]
	text = it.Summary
	if it.FromNode != "" {
		if t, ok := titles[it.FromNode]; ok {
			text = t + ": " + it.Summary
		}
	}
	return it.Date.Format(changelogDateLayout), text
}

func nodeTitles(g *graph.Graph) map[string]string {
	m := make(map[string]string, len(g.Nodes))
	for _, n := range g.Nodes {
		m[n.ID] = n.Title
	}
	return m
}

func formatDate(t time.Time) string { return t.Format(changelogDateLayout) }
