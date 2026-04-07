// Package templates はロードマッパーのプロジェクトテンプレートを embed で提供する。
package templates

import "embed"

//go:embed all:data
var FS embed.FS
