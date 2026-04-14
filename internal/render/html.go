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
	Title         string        `json:"title"`
	HTML          string        `json:"html"`
	Type          string        `json:"type,omitempty"`
	Links         []config.Link `json:"links,omitempty"`
	Parents       []string      `json:"parents,omitempty"`
	Children      []string      `json:"children,omitempty"`
	Difficulty    string        `json:"difficulty,omitempty"`
	EstimatedTime string        `json:"estimatedTime,omitempty"`
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

	nodeMeta, nodeOrder := buildNodeMeta(g, nodeHTML)
	nodeDataJSON, err := json.Marshal(nodeMeta)
	if err != nil {
		return "", err
	}
	nodeOrderJSON, err := json.Marshal(nodeOrder)
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
		"NodeOrderJSON":   template.JS(nodeOrderJSON),
		"HasMermaid":      hasMermaid,
		"OGPUrl":          ogpURL,
		"ChromaCSS":       template.CSS(ChromaCSS()),
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

	nodeIds := map[string][]string{}
	for rmID, g := range graphs {
		nodeIds[rmID] = graphNodeOrder(g)
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

// graphNodeOrder は g.Nodes の DAG 順序で required ノードの ID スライスを返す。
// optional / alternative ノードは進捗の分母に含めないためここで除外する。
func graphNodeOrder(g *graph.Graph) []string {
	order := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.Node.Type == config.NodeTypeOptional || n.Node.Type == config.NodeTypeAlternative {
			continue
		}
		order = append(order, n.ID)
	}
	return order
}

// buildNodeMeta は g.Nodes を1パスでメタデータマップと DAG 順序スライスを返す。
// nodeHTML が nil の場合は HTML フィールドを空にする。
func buildNodeMeta(g *graph.Graph, nodeHTML map[string]string) (map[string]NodeMeta, []string) {
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
			Type:          string(n.Node.Type),
			Links:         n.Node.Links,
			Parents:       parentIDs,
			Children:      childIDs,
			Difficulty:    string(n.Node.Difficulty),
			EstimatedTime: n.Node.EstimatedTime,
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
