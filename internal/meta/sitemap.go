package meta

import (
	"encoding/xml"
	"strings"

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
	siteURL := strings.TrimRight(cfg.Site.SiteURL, "/")
	if siteURL == "" {
		return "", nil
	}
	// basePath は "/" 始まりで "/" 終わりにする。空の場合は "/"
	basePath := cfg.Site.BasePath
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	base := siteURL + basePath

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
