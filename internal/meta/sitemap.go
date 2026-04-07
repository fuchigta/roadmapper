package meta

import (
	"encoding/xml"

	"github.com/fuchigta/roadmapper/internal/config"
)

type urlsetXML struct {
	XMLName xml.Name  `xml:"urlset"`
	Xmlns   string    `xml:"xmlns,attr"`
	URLs    []urlXML  `xml:"url"`
}

type urlXML struct {
	Loc string `xml:"loc"`
}

// RenderSitemap は sitemap.xml の文字列を返す。
// siteURL が空の場合は空文字列を返す。
func RenderSitemap(cfg *config.Config) (string, error) {
	base := SiteBase(cfg.Site.SiteURL, cfg.Site.BasePath)
	if base == "" {
		return "", nil
	}

	urls := []urlXML{
		{Loc: base},
	}
	for _, rm := range cfg.Roadmaps {
		urls = append(urls, urlXML{
			Loc: base + rm.ID + "/index.html",
		})
	}

	out, err := xml.MarshalIndent(urlsetXML{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(out), nil
}
