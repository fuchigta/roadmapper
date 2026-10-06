package render

import (
	"html/template"
	"strings"

	"github.com/fuchigta/roadmapper/internal/config"
)

// analyticsTmpl は provider ごとの <head> 断片。属性値は html/template が文脈に応じてエスケープする。
// アダプタは window.roadmapperTrack(name, data) を各サービスの API に橋渡しする。
var analyticsTmpl = template.Must(template.New("analytics").Parse(`
{{- define "umami" -}}
<script defer src="{{.ScriptURL}}" data-website-id="{{.SiteID}}"{{if .Domains}} data-domains="{{.Domains}}"{{end}}{{if .ExcludeSearch}} data-exclude-search="true"{{end}}></script>
{{if .Events}}<script>window.roadmapperTrack=function(n,d){try{window.umami&&umami.track(n,d)}catch(e){}};</script>
{{end}}{{end -}}
{{- define "plausible" -}}
<script defer data-domain="{{.SiteID}}" src="{{.ScriptURL}}"></script>
{{if .Events}}<script>window.roadmapperTrack=function(n,d){try{window.plausible&&plausible(n,{props:d})}catch(e){}};</script>
{{end}}{{end -}}
{{- define "goatcounter" -}}
<script data-goatcounter="{{.SiteID}}" async src="{{.ScriptURL}}"></script>
{{if .Events}}<script>window.roadmapperTrack=function(n,d){try{window.goatcounter&&goatcounter.count({path:n,title:JSON.stringify(d||{}),event:true})}catch(e){}};</script>
{{end}}{{end -}}
`))

// RenderAnalyticsHead は <head> に挿入するアクセス解析タグを返す。無効時は空。
// custom は Head をそのまま返す (設定者が信頼できる入力であることが前提)。
func RenderAnalyticsHead(a config.Analytics) template.HTML {
	switch a.Provider {
	case config.AnalyticsUmami, config.AnalyticsPlausible, config.AnalyticsGoatCounter:
		var sb strings.Builder
		data := map[string]any{
			"ScriptURL":     a.ScriptURL,
			"SiteID":        a.SiteID,
			"Domains":       strings.Join(a.Domains, ","),
			"ExcludeSearch": a.ExcludeSearchEnabled(),
			"Events":        a.EventsEnabled(),
		}
		if err := analyticsTmpl.ExecuteTemplate(&sb, a.Provider, data); err != nil {
			return ""
		}
		return template.HTML(sb.String())
	case config.AnalyticsCustom:
		return template.HTML(a.Head)
	}
	return ""
}

// analyticsEventsEnabled は SITE_CONFIG.analyticsEvents に渡す値 (解析有効 かつ events 有効)。
func analyticsEventsEnabled(a config.Analytics) bool {
	return a.Enabled() && a.EventsEnabled()
}
