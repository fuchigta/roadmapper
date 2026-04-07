package render

import (
	"encoding/json"
	"html/template"
	"io/fs"
	"strings"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/meta"
)

// NodeMeta はフロントエンドに渡すノードのメタデータ。
type NodeMeta struct {
	Title string        `json:"title"`
	HTML  string        `json:"html"`
	Links []config.Link `json:"links,omitempty"`
}

// RenderRoadmapPage は roadmap.html を使ってロードマップページの HTML を生成する。
func RenderRoadmapPage(
	webFS fs.FS,
	cfg *config.Config,
	rm *config.Roadmap,
	g *graph.Graph,
	lr *layout.Result,
	nodeHTML map[string]string, // nodeID → rendered HTML
	basePath string,
	assetBase string, // CSS/JS への相対パス (basePath 空なら "../")
	hasMermaid bool, // mermaid コードブロックがあれば mermaid.js を読み込む
) (string, error) {
	colors := DeriveColors(cfg.Site.BrandColor)

	// ノードデータ JSON
	nodeMeta := buildNodeMeta(g, nodeHTML)
	nodeDataJSON, err := json.Marshal(nodeMeta)
	if err != nil {
		return "", err
	}

	svgStr := RenderSVG(g, lr, cfg.Site.BrandColor)

	ogpURL := ""
	if base := meta.SiteBase(cfg.Site.SiteURL, basePath); base != "" {
		ogpURL = base + rm.ID + "/index.html"
	}

	tmplData := map[string]any{
		"Site":            cfg.Site,
		"Roadmap":         rm,
		"SVG":             template.HTML(svgStr),
		"BrandColor":      colors.Base,
		"BrandColorLight": colors.Light,
		"BasePath":        basePath,
		"AssetBase":       assetBase,
		"BasePathJSON":    jsonStr(basePath),
		"RepoJSON":        jsonStr(cfg.Site.Repo),
		"EditBranchJSON":  jsonStr(cfg.Site.EditBranch),
		"RoadmapIdJSON":   jsonStr(rm.ID),
		"NodeDataJSON":    template.JS(nodeDataJSON),
		"HasMermaid":      hasMermaid,
		"OGPUrl":          ogpURL,
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
) (string, error) {
	colors := DeriveColors(cfg.Site.BrandColor)

	ids := make([]string, len(cfg.Roadmaps))
	for i, rm := range cfg.Roadmaps {
		ids[i] = rm.ID
	}
	idsJSON, _ := json.Marshal(ids)

	// ロードマップごとのノードID一覧 (index ページでの進捗計算用)
	nodeIds := map[string][]string{}
	for rmID, g := range graphs {
		ns := make([]string, len(g.Nodes))
		for i, n := range g.Nodes {
			ns[i] = n.ID
		}
		nodeIds[rmID] = ns
	}
	nodeIdsJSON, _ := json.Marshal(nodeIds)

	ogpURL := ""
	rssURL := ""
	if base := meta.SiteBase(cfg.Site.SiteURL, basePath); base != "" {
		ogpURL = base
		rssURL = base + "feed.rss"
	}

	tmplData := map[string]any{
		"Site":            cfg.Site,
		"Roadmaps":        cfg.Roadmaps,
		"BrandColor":      colors.Base,
		"BrandColorLight": colors.Light,
		"BasePath":        basePath,
		"BasePathJSON":    jsonStr(basePath),
		"RoadmapIdsJSON":  template.JS(idsJSON),
		"NodeIdsJSON":     template.JS(nodeIdsJSON),
		"OGPUrl":          ogpURL,
		"RSSUrl":          rssURL,
	}

	return renderTemplate(webFS, "templates/index.html", tmplData)
}

func buildNodeMeta(g *graph.Graph, nodeHTML map[string]string) map[string]NodeMeta {
	meta := map[string]NodeMeta{}
	for _, n := range g.Nodes {
		meta[n.ID] = NodeMeta{
			Title: n.Title,
			HTML:  nodeHTML[n.ID],
			Links: n.Node.Links,
		}
	}
	return meta
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
