package config

import "testing"

func TestApplyDefaults_panel(t *testing.T) {
	tests := []struct {
		name string
		in   Panel
		want Panel
	}{
		{"全て未設定", Panel{}, Panel{520, 320, 960}},
		{"一部指定", Panel{Width: 700}, Panel{700, 320, 960}},
		{"全て指定", Panel{400, 300, 500}, Panel{400, 300, 500}},
		{"maxWidth のみ小さく指定", Panel{MaxWidth: 400}, Panel{400, 320, 400}},
		{"maxWidth が既定 minWidth 未満", Panel{MaxWidth: 300}, Panel{300, 300, 300}},
		{"minWidth のみ大きく指定", Panel{MinWidth: 700}, Panel{700, 700, 960}},
		{"width のみ既定上限超え", Panel{Width: 1200}, Panel{1200, 320, 1200}},
		{"width のみ既定下限未満", Panel{Width: 280}, Panel{280, 280, 960}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Site: Site{Panel: tt.in}}
			applyDefaults(cfg)
			if cfg.Site.Panel != tt.want {
				t.Errorf("got %+v, want %+v", cfg.Site.Panel, tt.want)
			}
		})
	}
}

func TestApplyDefaults_analytics(t *testing.T) {
	f := false
	tests := []struct {
		name       string
		in         Analytics
		wantEvents bool
		wantExcl   bool
	}{
		{"未指定は両方 true", Analytics{}, true, true},
		{"明示 false を保持", Analytics{Events: &f, ExcludeSearch: &f}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Site: Site{Analytics: tt.in}}
			applyDefaults(cfg)
			a := cfg.Site.Analytics
			if a.Events == nil || a.ExcludeSearch == nil {
				t.Fatal("nil が補完されていない")
			}
			if a.EventsEnabled() != tt.wantEvents || a.ExcludeSearchEnabled() != tt.wantExcl {
				t.Errorf("events=%v exclude=%v", a.EventsEnabled(), a.ExcludeSearchEnabled())
			}
		})
	}
}
