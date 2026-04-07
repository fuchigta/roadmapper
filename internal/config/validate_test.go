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
