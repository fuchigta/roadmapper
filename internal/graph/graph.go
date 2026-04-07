// Package graph はロードマップのノード/エッジモデルと DAG 検証を提供する。
package graph

import (
	"fmt"

	"github.com/fuchigta/roadmapper/internal/config"
)

// Node はグラフ上の1ノードを表す。
type Node struct {
	*config.Node
	ParentNodes   []*Node
	ChildrenNodes []*Node
}

// Graph はロードマップ全体のノードグラフを表す。
type Graph struct {
	Nodes   []*Node
	NodeMap map[string]*Node
}

// Build は config.Roadmap からフラット化されたグラフを構築し、DAG 検証を行う。
func Build(rm *config.Roadmap) (*Graph, error) {
	g := &Graph{NodeMap: map[string]*Node{}}

	// 全ノードをフラット化 (children が循環していても安全)
	visited := map[string]bool{}
	flattenNodes(rm.Nodes, g, visited)

	// 親子リンクを張る
	visited = map[string]bool{}
	if err := linkEdges(rm.Nodes, g, visited); err != nil {
		return nil, err
	}

	// 循環検出
	if err := detectCycles(g); err != nil {
		return nil, err
	}

	return g, nil
}

func flattenNodes(nodes []*config.Node, g *Graph, visited map[string]bool) {
	for _, n := range nodes {
		if visited[n.ID] {
			continue
		}
		visited[n.ID] = true
		gn := &Node{Node: n}
		g.Nodes = append(g.Nodes, gn)
		g.NodeMap[n.ID] = gn
		flattenNodes(n.Children, g, visited)
	}
}

func linkEdges(nodes []*config.Node, g *Graph, visited map[string]bool) error {
	for _, n := range nodes {
		if visited[n.ID] {
			continue
		}
		visited[n.ID] = true

		gn := g.NodeMap[n.ID]

		// 子→親
		for _, child := range n.Children {
			gc := g.NodeMap[child.ID]
			if gc == nil {
				continue
			}
			if !containsNode(gc.ParentNodes, gn) {
				gc.ParentNodes = append(gc.ParentNodes, gn)
			}
			if !containsNode(gn.ChildrenNodes, gc) {
				gn.ChildrenNodes = append(gn.ChildrenNodes, gc)
			}
		}

		// parents フィールド (複数親 DAG)
		for _, pid := range n.Parents {
			parent, ok := g.NodeMap[pid]
			if !ok {
				return fmt.Errorf("ノード %q: 親 %q が存在しません", n.ID, pid)
			}
			if !containsNode(gn.ParentNodes, parent) {
				gn.ParentNodes = append(gn.ParentNodes, parent)
			}
			if !containsNode(parent.ChildrenNodes, gn) {
				parent.ChildrenNodes = append(parent.ChildrenNodes, gn)
			}
		}

		if err := linkEdges(n.Children, g, visited); err != nil {
			return err
		}
	}
	return nil
}

func containsNode(ns []*Node, target *Node) bool {
	for _, n := range ns {
		if n.ID == target.ID {
			return true
		}
	}
	return false
}

// detectCycles は DFS で循環参照を検出する。
func detectCycles(g *Graph) error {
	visited := map[string]bool{}
	inStack := map[string]bool{}

	var dfs func(n *Node) error
	dfs = func(n *Node) error {
		visited[n.ID] = true
		inStack[n.ID] = true
		for _, child := range n.ChildrenNodes {
			if !visited[child.ID] {
				if err := dfs(child); err != nil {
					return err
				}
			} else if inStack[child.ID] {
				return fmt.Errorf("循環参照を検出しました: %q → %q", n.ID, child.ID)
			}
		}
		inStack[n.ID] = false
		return nil
	}

	for _, n := range g.Nodes {
		if !visited[n.ID] {
			if err := dfs(n); err != nil {
				return err
			}
		}
	}
	return nil
}

// Roots はグラフの根ノード (親を持たないノード) を返す。
func (g *Graph) Roots() []*Node {
	var roots []*Node
	for _, n := range g.Nodes {
		if len(n.ParentNodes) == 0 {
			roots = append(roots, n)
		}
	}
	return roots
}
