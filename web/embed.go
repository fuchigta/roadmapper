// Package web は roadmapper の Web アセット (HTML テンプレート・CSS・JS) を embed で提供する。
package web

import "embed"

//go:embed templates static
var FS embed.FS
