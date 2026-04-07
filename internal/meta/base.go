package meta

import "strings"

// SiteBase は siteURL と basePath から末尾スラッシュ付きのベース URL を返す。
// siteURL が空の場合は空文字列を返す。
func SiteBase(siteURL, basePath string) string {
	if siteURL == "" {
		return ""
	}
	if !strings.HasSuffix(basePath, "/") {
		basePath += "/"
	}
	return strings.TrimRight(siteURL, "/") + basePath
}
