package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
)

func NewValidateCmd() *cobra.Command {
	var (
		configPath string
		strict     bool
	)

	cmd := &cobra.Command{
		Use:           "validate",
		Short:         "roadmap.yml と content/ の整合性を検証する",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(configPath, strict)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "roadmap.yml",
		"設定ファイルのパス")
	cmd.Flags().BoolVar(&strict, "strict", false,
		"content/*.md が見つからないノードがあればエラーで終了する")

	return cmd
}

func runValidate(configPath string, strict bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("バリデーションに失敗しました")
	}

	// グラフ構造の検証 (循環参照など)
	graphs := make([]*graph.Graph, 0, len(cfg.Roadmaps))
	for i := range cfg.Roadmaps {
		rm := &cfg.Roadmaps[i]
		g, err := graph.Build(rm)
		if err != nil {
			return fmt.Errorf("ロードマップ %q: %w", rm.ID, err)
		}
		graphs = append(graphs, g)
	}

	// content/ の解決検証
	configDir := filepath.Dir(configPath)
	contentDir := filepath.Join(configDir, "content")
	docs, err := content.LoadDir(contentDir)
	if err != nil {
		return fmt.Errorf("content ディレクトリの読み込みに失敗: %w", err)
	}

	type unresolvedEntry struct {
		roadmapID string
		nodeID    string
		content   string
	}
	var unresolved []unresolvedEntry
	for i, g := range graphs {
		for _, n := range g.Nodes {
			if _, ok := lookupDoc(docs, n.Node); !ok {
				unresolved = append(unresolved, unresolvedEntry{
					roadmapID: cfg.Roadmaps[i].ID,
					nodeID:    n.ID,
					content:   n.Node.Content,
				})
			}
		}
	}

	if len(unresolved) > 0 {
		fmt.Fprintf(os.Stderr, "warning: %d 個のノードに対応する content/*.md が見つかりません:\n", len(unresolved))
		for _, e := range unresolved {
			if e.content != "" {
				fmt.Fprintf(os.Stderr, "  - [%s] %s (content: %q)\n", e.roadmapID, e.nodeID, e.content)
			} else {
				fmt.Fprintf(os.Stderr, "  - [%s] %s\n", e.roadmapID, e.nodeID)
			}
		}
		if strict {
			return fmt.Errorf("--strict 指定のため content 未解決を理由に失敗します")
		}
	}

	fmt.Printf("✓ %s の検証が完了しました (%d ロードマップ, content 解決済み %d / %d ノード)\n",
		configPath, len(cfg.Roadmaps), totalNodes(graphs)-len(unresolved), totalNodes(graphs))
	return nil
}

func totalNodes(graphs []*graph.Graph) int {
	n := 0
	for _, g := range graphs {
		n += len(g.Nodes)
	}
	return n
}
