package meta

import (
	"encoding/xml"
	"strings"
	"time"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
)

type rssXML struct {
	XMLName xml.Name    `xml:"rss"`
	Version string      `xml:"version,attr"`
	Channel channelXML  `xml:"channel"`
}

type channelXML struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	Items       []itemXML `xml:"item"`
}

type itemXML struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
}

// RenderRSS は RSS 2.0 フィードの文字列を返す。
// siteURL が空の場合は空文字列を返す。
func RenderRSS(cfg *config.Config, graphs map[string]*graph.Graph) (string, error) {
	siteURL := strings.TrimRight(cfg.Site.SiteURL, "/")
	if siteURL == "" {
		return "", nil
	}
	basePath := cfg.Site.BasePath
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	siteLink := siteURL + basePath

	var items []itemXML
	for _, rm := range cfg.Roadmaps {
		g, ok := graphs[rm.ID]
		if !ok {
			continue
		}
		for _, n := range g.Nodes {
			link := siteLink + rm.ID + "/index.html#" + n.ID
			items = append(items, itemXML{
				Title:       n.Title,
				Link:        link,
				Description: rm.Title + " — " + n.Title,
				GUID:        link,
			})
		}
	}

	feed := rssXML{
		Version: "2.0",
		Channel: channelXML{
			Title:       cfg.Site.Title,
			Link:        siteLink,
			Description: cfg.Site.Description,
			PubDate:     time.Now().Format(time.RFC1123Z),
			Items:       items,
		},
	}

	out, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}
