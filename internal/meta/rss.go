package meta

import (
	"encoding/xml"
	"fmt"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
)

type rssXML struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel channelXML `xml:"channel"`
}

type channelXML struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	LastBuild   string    `xml:"lastBuildDate"`
	Items       []itemXML `xml:"item"`
}

type itemXML struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate,omitempty"`
}

// RenderRSS は RSS 2.0 フィードの文字列を返す。
// siteURL が空の場合は空文字列を返す。
// logs[roadmapID] に改版履歴があるロードマップはその項目を item にし、
// なければノードごとの item にする。logs は nil でもよい。
func RenderRSS(cfg *config.Config, graphs map[string]*graph.Graph, logs map[string]*changelog.Log) (string, error) {
	siteLink := SiteBase(cfg.Site.SiteURL, cfg.Site.BasePath)
	if siteLink == "" {
		return "", nil
	}

	var items []itemXML
	var latest time.Time
	for _, rm := range cfg.Roadmaps {
		g, ok := graphs[rm.ID]
		if !ok {
			continue
		}
		pageURL := siteLink + rm.ID + "/index.html"
		if log := logs[rm.ID]; log != nil && len(log.Items) > 0 {
			for i, it := range log.Items {
				items = append(items, changelogItem(rm, g, pageURL, it, i))
			}
			if log.Latest.After(latest) {
				latest = log.Latest
			}
			continue
		}
		for _, n := range g.Nodes {
			link := pageURL + "#" + n.ID
			items = append(items, itemXML{
				Title:       n.Title,
				Link:        link,
				Description: rm.Title + " — " + n.Title,
				GUID:        link,
			})
		}
	}

	// 改版履歴があれば最新日、なければビルド時刻
	pub := time.Now()
	if !latest.IsZero() {
		pub = latest
	}
	feed := rssXML{
		Version: "2.0",
		Channel: channelXML{
			Title:       cfg.Site.Title,
			Link:        siteLink,
			Description: cfg.Site.Description,
			PubDate:     pub.Format(time.RFC1123Z),
			LastBuild:   pub.Format(time.RFC1123Z),
			Items:       items,
		},
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}

// changelogItem は改版履歴 1 項目を RSS item に変換する。idx は Log.Items 内の位置 (guid の一意化に使う)。
func changelogItem(rm config.Roadmap, g *graph.Graph, pageURL string, it changelog.Item, idx int) itemXML {
	title := it.Summary
	link := pageURL
	nodeID := it.FromNode
	if nodeID == "" && len(it.NodeIDs) == 1 {
		nodeID = it.NodeIDs[0]
	}
	if nodeID != "" {
		link += "#" + nodeID
	}
	if it.FromNode != "" {
		if n, ok := g.NodeMap[it.FromNode]; ok {
			title = n.Title + ": " + it.Summary
		}
	}
	return itemXML{
		Title:       title,
		Link:        link,
		Description: rm.Title + " — " + it.Summary,
		GUID:        fmt.Sprintf("%s#changelog-%s-%d", pageURL, it.Date.Format(config.ChangelogDateLayout), idx),
		PubDate:     it.Date.Format(time.RFC1123Z),
	}
}
