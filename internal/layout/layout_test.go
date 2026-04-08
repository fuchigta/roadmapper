package layout_test

import (
	"strings"
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
)

func makeConfig() *config.Config {
	return &config.Config{
		Site: config.Site{
			BrandColor: "#4f46e5",
			Layout: config.Layout{
				RankDir: "TB",
				NodeSep: 50,
				RankSep: 80,
			},
		},
	}
}

func TestWrapTitle(t *testing.T) {
	const (
		testFontSize = 13.0
		testMaxInner = 260.0 - 24.0 // maxNodeWidth - NodePaddingX
	)
	tests := []struct {
		name      string
		title     string
		wantLines int
		wantTrunc bool // 最終行が … を含む
	}{
		{"short ascii", "Go", 1, false},
		{"long ascii with spaces", "Introduction to Frontend Development Basics", 2, false},
		{"long japanese no spaces", "フロントエンド開発の基礎と応用について学ぶ", 2, false},
		{"3 lines no truncation", "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめも", 3, false},
		{"very long truncated", "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやゆよらりるれろわをんアイウエオカキクケコサシスセソ", 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := layout.WrapTitle(tt.title, testFontSize, testMaxInner)
			if len(lines) != tt.wantLines {
				t.Errorf("WrapTitle(%q): got %d lines, want %d", tt.title, len(lines), tt.wantLines)
			}
			if tt.wantTrunc {
				last := lines[len(lines)-1]
				if !strings.HasSuffix(last, "…") {
					t.Errorf("WrapTitle(%q): last line %q should end with …", tt.title, last)
				}
			}
		})
	}
}

func TestNodeSize(t *testing.T) {
	const (
		minW = 180.0
		minH = 50.0
		maxW = 260.0
	)
	t.Run("short title stays at minNodeWidth", func(t *testing.T) {
		w, h := layout.NodeSize("Go")
		if w != minW {
			t.Errorf("want width=%v, got %v", minW, w)
		}
		if h != minH {
			t.Errorf("want height=%v, got %v", minH, h)
		}
	})
	t.Run("long title increases width or height", func(t *testing.T) {
		w, h := layout.NodeSize("フロントエンド開発の基礎と応用について")
		if w <= minW && h <= minH {
			t.Errorf("long title should increase width or height: w=%v h=%v", w, h)
		}
		if w > maxW {
			t.Errorf("width %v exceeds maxNodeWidth %v", w, maxW)
		}
	})
}

func TestCompute_basic(t *testing.T) {
	rm := &config.Roadmap{
		ID: "test", Title: "Test",
		Nodes: []*config.Node{
			{
				ID: "a", Title: "A", Type: config.NodeTypeRequired,
				Children: []*config.Node{
					{ID: "b", Title: "B", Type: config.NodeTypeRequired},
					{ID: "c", Title: "C", Type: config.NodeTypeOptional},
				},
			},
		},
	}

	g, err := graph.Build(rm)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	cfg := makeConfig()
	result, err := layout.Compute(g, cfg)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	if len(result.Nodes) != 3 {
		t.Errorf("want 3 nodes in result, got %d", len(result.Nodes))
	}

	for id, nl := range result.Nodes {
		if nl.Width == 0 || nl.Height == 0 {
			t.Errorf("node %q has zero width/height", id)
		}
		if nl.X == 0 && nl.Y == 0 && id != "a" {
			t.Errorf("node %q appears to be at origin (layout may have failed)", id)
		}
	}

	// dagre は TB レイアウトで a が b, c より上になるはず
	aY := result.Nodes["a"].Y
	bY := result.Nodes["b"].Y
	if aY >= bY {
		t.Errorf("TB layout: node a (y=%v) should be above node b (y=%v)", aY, bY)
	}
}

func TestCompute_manualOverride(t *testing.T) {
	x, y := 500.0, 200.0
	rm := &config.Roadmap{
		ID: "test", Title: "Test",
		Nodes: []*config.Node{
			{ID: "a", Title: "A", Type: config.NodeTypeRequired, X: &x, Y: &y},
			{ID: "b", Title: "B", Type: config.NodeTypeRequired},
		},
	}

	g, err := graph.Build(rm)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	result, err := layout.Compute(g, makeConfig())
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}

	na := result.Nodes["a"]
	if na.X != 500 || na.Y != 200 {
		t.Errorf("manual override not applied: got x=%v y=%v", na.X, na.Y)
	}
}
