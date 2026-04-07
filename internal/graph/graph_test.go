package graph_test

import (
	"testing"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
)

func makeRoadmap(nodes []*config.Node) *config.Roadmap {
	return &config.Roadmap{ID: "test", Title: "Test", Nodes: nodes}
}

func TestBuild_simple(t *testing.T) {
	rm := makeRoadmap([]*config.Node{
		{
			ID: "a", Title: "A", Type: config.NodeTypeRequired,
			Children: []*config.Node{
				{ID: "b", Title: "B", Type: config.NodeTypeOptional},
			},
		},
	})

	g, err := graph.Build(rm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) != 2 {
		t.Errorf("want 2 nodes, got %d", len(g.Nodes))
	}

	nodeA := g.NodeMap["a"]
	if len(nodeA.ChildrenNodes) != 1 {
		t.Errorf("node a should have 1 child")
	}
}

func TestBuild_dag_multipleParents(t *testing.T) {
	rm := makeRoadmap([]*config.Node{
		{ID: "a", Title: "A", Type: config.NodeTypeRequired},
		{ID: "b", Title: "B", Type: config.NodeTypeRequired},
		{ID: "c", Title: "C", Type: config.NodeTypeRequired, Parents: []string{"a", "b"}},
	})

	g, err := graph.Build(rm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodeC := g.NodeMap["c"]
	if len(nodeC.ParentNodes) != 2 {
		t.Errorf("node c should have 2 parents, got %d", len(nodeC.ParentNodes))
	}
}

func TestBuild_detectCycle(t *testing.T) {
	// a → b → c → a は循環
	nodeA := &config.Node{ID: "a", Title: "A", Type: config.NodeTypeRequired}
	nodeB := &config.Node{ID: "b", Title: "B", Type: config.NodeTypeRequired}
	nodeC := &config.Node{ID: "c", Title: "C", Type: config.NodeTypeRequired}
	nodeA.Children = []*config.Node{nodeB}
	nodeB.Children = []*config.Node{nodeC}
	nodeC.Children = []*config.Node{nodeA}

	rm := makeRoadmap([]*config.Node{nodeA})
	if _, err := graph.Build(rm); err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestGraph_Roots(t *testing.T) {
	rm := makeRoadmap([]*config.Node{
		{
			ID: "root", Title: "Root", Type: config.NodeTypeRequired,
			Children: []*config.Node{
				{ID: "child", Title: "Child", Type: config.NodeTypeRequired},
			},
		},
	})

	g, err := graph.Build(rm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	roots := g.Roots()
	if len(roots) != 1 || roots[0].ID != "root" {
		t.Errorf("expected 1 root with id 'root', got %v", roots)
	}
}
