package templates_test

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/templates"
)

// TestTemplates_strictValid は全テンプレートが `validate --strict` 相当の検証を
// 警告なしで通ること (全ノードの content 解決・改版履歴の整合) を確認する。
func TestTemplates_strictValid(t *testing.T) {
	entries, err := fs.ReadDir(templates.FS, "data")
	if err != nil {
		t.Fatalf("テンプレート一覧の取得に失敗: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("data", name)
			cfg, err := config.Load(filepath.Join(dir, "roadmap.yml"))
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			if err := config.Validate(cfg); err != nil {
				t.Fatalf("config.Validate: %v", err)
			}
			docs, err := content.LoadDir(filepath.Join(dir, "content"))
			if err != nil {
				t.Fatalf("content.LoadDir: %v", err)
			}
			for i := range cfg.Roadmaps {
				rm := &cfg.Roadmaps[i]
				g, err := graph.Build(rm)
				if err != nil {
					t.Fatalf("[%s] graph.Build: %v", rm.ID, err)
				}
				nodeDocs := map[string]*content.Doc{}
				for _, n := range g.Nodes {
					key := n.ID
					if n.Node.Content != "" {
						key = n.Node.Content
					}
					doc, ok := docs[key]
					if !ok {
						t.Errorf("[%s] ノード %q の content が見つかりません", rm.ID, n.ID)
						continue
					}
					nodeDocs[n.ID] = doc
				}
				if _, err := changelog.Build(rm.Changelog, nodeDocs); err != nil {
					t.Errorf("[%s] changelog.Build: %v", rm.ID, err)
				}
				for _, w := range changelog.Check(rm.Changelog, nodeDocs, time.Now()) {
					t.Errorf("[%s] 改版履歴の警告: %s %s", rm.ID, w.NodeID, w.Message)
				}
			}
		})
	}
}
