package render

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"strings"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/meta"
)

// NodeMeta はフロントエンドに渡すノードのメタデータ。
type NodeMeta struct {
	Title         string        `json:"title"`
	HTML          string        `json:"html"`
	Text          string        `json:"text,omitempty"`
	Type          string        `json:"type,omitempty"`
	Links         []config.Link `json:"links,omitempty"`
	Parents       []string      `json:"parents,omitempty"`
	Children      []string      `json:"children,omitempty"`
	Difficulty    string        `json:"difficulty,omitempty"`
	EstimatedTime string        `json:"estimatedTime,omitempty"`
	// EditPath は記事ファイルのリポジトリルート相対パス (`/` 区切り)。「この記事を編集」リンク用。
	EditPath string `json:"editPath,omitempty"`
	// Draft は下書きノード。公開ビルドでは html / text / links / editPath を空にして draft のみを渡す (本文を漏らさない)。
	Draft bool `json:"draft,omitempty"`
}

// RenderRoadmapPage は roadmap.html を使ってロードマップページの HTML を生成する。
func RenderRoadmapPage(
	webFS fs.FS,
	cfg *config.Config,
	rm *config.Roadmap,
	g *graph.Graph,
	lr *layout.Result,
	nodeHTML map[string]string, // nodeID → rendered HTML
	nodeText map[string]string, // nodeID → plaintext (全文検索用)
	basePath string,
	assetBase string, // CSS/JS への相対パス (basePath 空なら "../")
	hasMermaid bool, // mermaid コードブロックがあれば mermaid.js を読み込む
	log *changelog.Log, // 改版履歴 (nil または空なら履歴 UI を出さない)
	badges map[string]time.Time, // 更新バッジを付けるノード ID → 更新日 (nil 可)
	editPaths map[string]string, // ノード ID → 記事のリポジトリルート相対パス (nil 可)
	drafts Drafts, // 下書きノードの扱い (ゼロ値なら下書きなし)
) (string, error) {
	colors := DeriveColors(cfg.Site.BrandColor)

	titles := nodeTitles(g)
	hasLog := log != nil && len(log.Items) > 0
	if hasLog {
		// ノード記事の末尾に履歴セクションを静的に付与する (呼び出し元の map は変更しない)
		withHist := make(map[string]string, len(nodeHTML))
		for id, h := range nodeHTML {
			withHist[id] = h
		}
		for _, n := range g.Nodes {
			if drafts.locked(n.ID) {
				continue
			}
			if sec := renderNodeHistory(log, n.ID, titles); sec != "" {
				withHist[n.ID] += sec
			}
		}
		nodeHTML = withHist
	}

	nodeMeta, nodeOrder := buildNodeMeta(g, nodeHTML, nodeText, editPaths, drafts)
	nodeDataJSON, err := json.Marshal(nodeMeta)
	if err != nil {
		return "", err
	}
	nodeOrderJSON, err := json.Marshal(nodeOrder)
	if err != nil {
		return "", err
	}

	svgStr := RenderSVGWithBadges(g, lr, cfg.Site.BrandColor, badges, drafts)

	ogpURL := ""
	if base := meta.SiteBase(cfg.Site.SiteURL, basePath); base != "" {
		ogpURL = base + rm.ID + "/index.html"
	}

	tmplData := map[string]any{
		"HasChangelog":     hasLog,
		"Site":             cfg.Site,
		"Roadmap":          rm,
		"SVG":              template.HTML(svgStr),
		"BrandColor":       colors.Base,
		"BrandColorLight":  colors.Light,
		"BasePath":         basePath,
		"AssetBase":        assetBase,
		"BasePathJSON":     jsonStr(basePath),
		"RepoJSON":         jsonStr(cfg.Site.Repo),
		"EditBranchJSON":   jsonStr(cfg.Site.EditBranch),
		"RoadmapIdJSON":    jsonStr(rm.ID),
		"ProgressSyncJSON": progressSyncJSON(cfg),
		"AnalyticsHead":    RenderAnalyticsHead(cfg.Site.Analytics),
		"AnalyticsEvents":  analyticsEventsEnabled(cfg.Site.Analytics),
		"NodeDataJSON":     template.JS(nodeDataJSON),
		"NodeOrderJSON":    template.JS(nodeOrderJSON),
		"HasMermaid":       hasMermaid,
		"DraftPreview":     drafts.Preview,
		"OGPUrl":           ogpURL,
		"ChromaCSS":        template.CSS(ChromaCSS()),
	}

	if hasLog {
		tmplData["LatestDate"], tmplData["LatestText"] = latestLine(log, titles)
		tmplData["PanelHistory"] = renderPanelHistory(log, titles)
	}

	return renderTemplate(webFS, "templates/roadmap.html", tmplData)
}

// RenderIndexPage はサイトのトップページを生成する。
// graphs は ロードマップID → グラフ のマップ。カード進捗計算用のノードID一覧に使う。
func RenderIndexPage(
	webFS fs.FS,
	cfg *config.Config,
	basePath string,
	graphs map[string]*graph.Graph,
	latest map[string]time.Time, // ロードマップID → 最新更新日 (zero または未登録なら表示しない)
	locked map[string]map[string]bool, // ロードマップ ID → 進捗の分母から除く非公開 (下書き) ノード ID 集合 (nil 可)
) (string, error) {
	colors := DeriveColors(cfg.Site.BrandColor)

	ids := make([]string, len(cfg.Roadmaps))
	for i, rm := range cfg.Roadmaps {
		ids[i] = rm.ID
	}
	idsJSON, _ := json.Marshal(ids)

	nodeIds := map[string][]string{}
	for rmID, g := range graphs {
		nodeIds[rmID] = graphNodeOrder(g, locked[rmID])
	}
	nodeIdsJSON, _ := json.Marshal(nodeIds)

	ogpURL := ""
	rssURL := ""
	if base := meta.SiteBase(cfg.Site.SiteURL, basePath); base != "" {
		ogpURL = base
		rssURL = base + "feed.rss"
	}

	latestStr := map[string]string{}
	for id, t := range latest {
		if !t.IsZero() {
			latestStr[id] = formatDate(t)
		}
	}

	tmplData := map[string]any{
		"Latest":           latestStr,
		"Site":             cfg.Site,
		"Roadmaps":         cfg.Roadmaps,
		"BrandColor":       colors.Base,
		"BrandColorLight":  colors.Light,
		"BasePath":         basePath,
		"BasePathJSON":     jsonStr(basePath),
		"ProgressSyncJSON": progressSyncJSON(cfg),
		"AnalyticsHead":    RenderAnalyticsHead(cfg.Site.Analytics),
		"AnalyticsEvents":  analyticsEventsEnabled(cfg.Site.Analytics),
		"RoadmapIdsJSON":   template.JS(idsJSON),
		"NodeIdsJSON":      template.JS(nodeIdsJSON),
		"OGPUrl":           ogpURL,
		"RSSUrl":           rssURL,
	}

	return renderTemplate(webFS, "templates/index.html", tmplData)
}

// graphNodeOrder は g.Nodes の DAG 順序で required ノードの ID スライスを返す。
// optional / alternative ノードは進捗の分母に含めないためここで除外する。
// exclude に含まれるノード (非公開の下書き) も除外する。
func graphNodeOrder(g *graph.Graph, exclude map[string]bool) []string {
	order := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.Node.Type == config.NodeTypeOptional || n.Node.Type == config.NodeTypeAlternative || exclude[n.ID] {
			continue
		}
		order = append(order, n.ID)
	}
	return order
}

// buildNodeMeta は g.Nodes を1パスでメタデータマップと DAG 順序スライスを返す。
// nodeHTML / nodeText が nil の場合は対応フィールドを空にする。
func buildNodeMeta(g *graph.Graph, nodeHTML, nodeText, editPaths map[string]string, drafts Drafts) (map[string]NodeMeta, []string) {
	meta := make(map[string]NodeMeta, len(g.Nodes))
	order := make([]string, len(g.Nodes))
	for i, n := range g.Nodes {
		parentIDs := make([]string, len(n.ParentNodes))
		for j, p := range n.ParentNodes {
			parentIDs[j] = p.ID
		}
		childIDs := make([]string, len(n.ChildrenNodes))
		for j, c := range n.ChildrenNodes {
			childIDs[j] = c.ID
		}
		meta[n.ID] = NodeMeta{
			Title:         n.Title,
			HTML:          nodeHTML[n.ID],
			Text:          nodeText[n.ID],
			Type:          string(n.Node.Type),
			Links:         n.Node.Links,
			Parents:       parentIDs,
			Children:      childIDs,
			Difficulty:    string(n.Node.Difficulty),
			EstimatedTime: n.Node.EstimatedTime,
			EditPath:      editPaths[n.ID],
		}
		if drafts.IDs[n.ID] {
			m := meta[n.ID]
			m.Draft = true
			if drafts.locked(n.ID) {
				m.HTML, m.Text, m.Links, m.EditPath = "", "", nil, ""
			}
			meta[n.ID] = m
		}
		order[i] = n.ID
	}
	return meta, order
}

func renderTemplate(webFS fs.FS, name string, data any) (string, error) {
	tmplBytes, err := fs.ReadFile(webFS, name)
	if err != nil {
		return "", err
	}

	t, err := template.New(name).Parse(string(tmplBytes))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	if err := t.Execute(&sb, data); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func jsonStr(s string) template.JS {
	b, _ := json.Marshal(s)
	return template.JS(b)
}

func progressSyncJSON(cfg *config.Config) template.JS {
	b, _ := json.Marshal(map[string]any{
		"enabled":  cfg.Site.ProgressSync.Enabled,
		"endpoint": strings.TrimRight(cfg.Site.ProgressSync.Endpoint, "/"),
	})
	return template.JS(b)
}
