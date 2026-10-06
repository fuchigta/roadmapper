package config_test

import (
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
)

func TestValidate_valid(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{
			{
				ID:    "frontend",
				Title: "Frontend",
				Nodes: []*config.Node{
					{
						ID:    "html",
						Title: "HTML",
						Type:  config.NodeTypeRequired,
						Children: []*config.Node{
							{ID: "css", Title: "CSS", Type: config.NodeTypeRequired},
						},
					},
				},
			},
		},
	}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("valid config should not error: %v", err)
	}
}

func TestValidate_missingTitle(t *testing.T) {
	cfg := &config.Config{
		Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1"}},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error for missing site title")
	}
}

func TestValidate_duplicateNodeID(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{
			{
				ID:    "r1",
				Title: "R1",
				Nodes: []*config.Node{
					{ID: "dup", Title: "A", Type: config.NodeTypeRequired},
					{ID: "dup", Title: "B", Type: config.NodeTypeRequired},
				},
			},
		},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error for duplicate node ID")
	}
}

func TestValidate_validDifficulty(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{
			{
				ID:    "r1",
				Title: "R1",
				Nodes: []*config.Node{
					{ID: "a", Title: "A", Type: config.NodeTypeRequired, Difficulty: config.DifficultyBeginner},
					{ID: "b", Title: "B", Type: config.NodeTypeRequired, Difficulty: config.DifficultyIntermediate},
					{ID: "c", Title: "C", Type: config.NodeTypeRequired, Difficulty: config.DifficultyAdvanced},
				},
			},
		},
	}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("valid difficulty should not error: %v", err)
	}
}

func TestValidate_emptyDifficulty(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{
			{
				ID:    "r1",
				Title: "R1",
				Nodes: []*config.Node{
					{ID: "a", Title: "A", Type: config.NodeTypeRequired},
				},
			},
		},
	}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("empty difficulty should not error: %v", err)
	}
}

func TestValidate_invalidDifficulty(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{
			{
				ID:    "r1",
				Title: "R1",
				Nodes: []*config.Node{
					{ID: "a", Title: "A", Type: config.NodeTypeRequired, Difficulty: "expert"},
				},
			},
		},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error for invalid difficulty")
	}
}

func TestValidate_validProgressSync(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{
			Title:        "Test",
			ProgressSync: config.ProgressSync{Enabled: true, Endpoint: "https://api.example.com/sync"},
		},
		Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1", Nodes: []*config.Node{{ID: "a", Title: "A", Type: config.NodeTypeRequired}}}},
	}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("valid progressSync should not error: %v", err)
	}
}

func TestValidate_emptyProgressSync(t *testing.T) {
	cfg := &config.Config{
		Site:     config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1", Nodes: []*config.Node{{ID: "a", Title: "A", Type: config.NodeTypeRequired}}}},
	}
	if err := config.Validate(cfg); err != nil {
		t.Fatalf("disabled progressSync with empty endpoint should not error: %v", err)
	}
}

func TestValidate_invalidProgressSyncMissingEndpoint(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{
			Title:        "Test",
			ProgressSync: config.ProgressSync{Enabled: true, Endpoint: ""},
		},
		Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1", Nodes: []*config.Node{{ID: "a", Title: "A", Type: config.NodeTypeRequired}}}},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error for missing endpoint when enabled")
	}
}

func TestValidate_invalidProgressSyncBadScheme(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{
			Title:        "Test",
			ProgressSync: config.ProgressSync{Enabled: true, Endpoint: "ftp://example.com/sync"},
		},
		Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1", Nodes: []*config.Node{{ID: "a", Title: "A", Type: config.NodeTypeRequired}}}},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error for invalid scheme")
	}
}

func TestValidate_unknownParent(t *testing.T) {
	cfg := &config.Config{
		Site: config.Site{Title: "Test"},
		Roadmaps: []config.Roadmap{
			{
				ID:    "r1",
				Title: "R1",
				Nodes: []*config.Node{
					{ID: "a", Title: "A", Type: config.NodeTypeRequired, Parents: []string{"nonexistent"}},
				},
			},
		},
	}
	if err := config.Validate(cfg); err == nil {
		t.Fatal("expected error for unknown parent")
	}
}

func TestValidate_panel(t *testing.T) {
	tests := []struct {
		name    string
		panel   config.Panel
		wantErr bool
	}{
		{"未設定", config.Panel{}, false},
		{"既定値", config.Panel{Width: 520, MinWidth: 320, MaxWidth: 960}, false},
		{"境界値が等しい", config.Panel{Width: 400, MinWidth: 400, MaxWidth: 400}, false},
		{"負の width", config.Panel{Width: -1, MinWidth: 320, MaxWidth: 960}, true},
		{"負の minWidth", config.Panel{Width: 520, MinWidth: -320, MaxWidth: 960}, true},
		{"負の maxWidth", config.Panel{Width: 520, MinWidth: 320, MaxWidth: -1}, true},
		{"minWidth > width", config.Panel{Width: 300, MinWidth: 320, MaxWidth: 960}, true},
		{"width > maxWidth", config.Panel{Width: 1000, MinWidth: 320, MaxWidth: 960}, true},
		{"minWidth > maxWidth", config.Panel{Width: 0, MinWidth: 1000, MaxWidth: 960}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Site:     config.Site{Title: "T", Panel: tt.panel},
				Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1"}},
			}
			err := config.Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate_analytics(t *testing.T) {
	tests := []struct {
		name    string
		a       config.Analytics
		wantErr bool
	}{
		{"無効", config.Analytics{}, false},
		{"umami 正常", config.Analytics{Provider: "umami", ScriptURL: "https://u.example.com/script.js", SiteID: "x"}, false},
		{"plausible 正常", config.Analytics{Provider: "plausible", ScriptURL: "https://plausible.io/js/script.js", SiteID: "example.com"}, false},
		{"goatcounter 正常", config.Analytics{Provider: "goatcounter", ScriptURL: "http://gc.example.com/count.js", SiteID: "https://x.goatcounter.com/count"}, false},
		{"custom 正常", config.Analytics{Provider: "custom", Head: "<script></script>"}, false},
		{"provider 不正", config.Analytics{Provider: "ga", ScriptURL: "https://x", SiteID: "x"}, true},
		{"scriptUrl 欠落", config.Analytics{Provider: "umami", SiteID: "x"}, true},
		{"scriptUrl スキーム不正", config.Analytics{Provider: "umami", ScriptURL: "//x/s.js", SiteID: "x"}, true},
		{"siteId 欠落", config.Analytics{Provider: "plausible", ScriptURL: "https://x/s.js"}, true},
		{"custom で head 欠落", config.Analytics{Provider: "custom"}, true},
		{"custom で head 空白のみ", config.Analytics{Provider: "custom", Head: "  "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Site:     config.Site{Title: "T", Analytics: tt.a},
				Roadmaps: []config.Roadmap{{ID: "r1", Title: "R1"}},
			}
			err := config.Validate(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
