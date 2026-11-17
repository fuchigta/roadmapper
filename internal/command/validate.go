package command

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/changelog"
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
		"content/*.md が見つからないノードや改版履歴の警告があればエラーで終了する")

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
	graphs, err := buildGraphs(cfg)
	if err != nil {
		return err
	}

	// content/ の解決検証
	configDir := filepath.Dir(configPath)
	contentDir := filepath.Join(configDir, "content")
	docs, err := content.LoadDir(contentDir)
	if err != nil {
		return fmt.Errorf("content ディレクトリの読み込みに失敗: %w", err)
	}

	// 下書きノードは記事が未作成でもよい (警告しない)
	draftSets := make([]map[string]bool, len(graphs))
	for i, g := range graphs {
		draftSets[i] = resolveDrafts(g, docs)
	}
	unresolved := findUnresolved(cfg, graphs, docs, draftSets)
	printUnresolved(unresolved)

	// 改版履歴の検証 (日付パース失敗はエラー、整合性の問題は警告)
	changelogWarnings, err := checkChangelogs(cfg, graphs, docs, draftSets, time.Now())
	if err != nil {
		return err
	}
	// 全警告を出し終えてから --strict 判定する
	if strict && (len(unresolved) > 0 || changelogWarnings > 0) {
		return fmt.Errorf("--strict 指定のため警告 (content 未解決 %d 件, 改版履歴 %d 件) を理由に失敗します",
			len(unresolved), changelogWarnings)
	}

	fmt.Printf("✓ %s の検証が完了しました (%d ロードマップ, content 解決済み %d / %d ノード)\n",
		configPath, len(cfg.Roadmaps), totalNodes(graphs)-len(unresolved), totalNodes(graphs))
	return nil
}

// buildGraphs は全ロードマップのグラフを構築する (循環参照などの検証を兼ねる)。
func buildGraphs(cfg *config.Config) ([]*graph.Graph, error) {
	graphs := make([]*graph.Graph, 0, len(cfg.Roadmaps))
	for i := range cfg.Roadmaps {
		rm := &cfg.Roadmaps[i]
		g, err := graph.Build(rm)
		if err != nil {
			return nil, fmt.Errorf("ロードマップ %q: %w", rm.ID, err)
		}
		graphs = append(graphs, g)
	}
	return graphs, nil
}

// unresolvedEntry は記事 (content/*.md) が見つからないノード。
type unresolvedEntry struct {
	roadmapID string
	nodeID    string
	content   string
}

// findUnresolved は記事が見つからないノードを返す。draftSets[i] (graphs[i] の下書き) に含まれるノードは除く。
func findUnresolved(cfg *config.Config, graphs []*graph.Graph, docs map[string]*content.Doc, draftSets []map[string]bool) []unresolvedEntry {
	var unresolved []unresolvedEntry
	for i, g := range graphs {
		for _, n := range g.Nodes {
			if draftSets[i][n.ID] {
				continue
			}
			if _, ok := lookupDoc(docs, n.Node); !ok {
				unresolved = append(unresolved, unresolvedEntry{
					roadmapID: cfg.Roadmaps[i].ID,
					nodeID:    n.ID,
					content:   n.Node.Content,
				})
			}
		}
	}
	return unresolved
}

// printUnresolved は記事が見つからないノードを stderr に警告する。
func printUnresolved(unresolved []unresolvedEntry) {
	if len(unresolved) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "warning: %d 個のノードに対応する content/*.md が見つかりません:\n", len(unresolved))
	for _, e := range unresolved {
		if e.content != "" {
			fmt.Fprintf(os.Stderr, "  - [%s] %s (content: %q)\n", e.roadmapID, e.nodeID, e.content)
		} else {
			fmt.Fprintf(os.Stderr, "  - [%s] %s\n", e.roadmapID, e.nodeID)
		}
	}
}

// checkChangelogs は全ロードマップの改版履歴を検証し、警告を stderr に出して件数を返す。
// 日付のパース失敗はエラー (それまでのロードマップの警告は出力済み)。
func checkChangelogs(cfg *config.Config, graphs []*graph.Graph, docs map[string]*content.Doc, draftSets []map[string]bool, now time.Time) (int, error) {
	total := 0
	for i, g := range graphs {
		n, err := checkRoadmapChangelog(&cfg.Roadmaps[i], g, docs, draftSets[i], now)
		if err != nil {
			return 0, err
		}
		total += n
	}
	return total, nil
}

// checkRoadmapChangelog は 1 ロードマップの改版履歴を検証し、警告を stderr に出して件数を返す。
func checkRoadmapChangelog(rm *config.Roadmap, g *graph.Graph, docs map[string]*content.Doc, drafts map[string]bool, now time.Time) (int, error) {
	// 公開ビルドと同じく、下書きの frontmatter と changelog の下書き参照は検査対象から外す
	entries := changelog.StripNodes(rm.Changelog, drafts)
	nodeDocs := resolveNodeDocs(g, docs, drafts)
	if _, err := changelog.Build(entries, nodeDocs); err != nil {
		return 0, fmt.Errorf("ロードマップ %q の改版履歴が不正: %w", rm.ID, err)
	}
	ws := changelog.Check(entries, nodeDocs, now)
	ws = append(ws, changelog.CheckDrafts(rm.Changelog, drafts)...)
	printChangelogWarnings(rm.ID, ws)
	return len(ws), nil
}

// printChangelogWarnings は改版履歴の警告を stderr に出す。警告がなければ何もしない。
func printChangelogWarnings(roadmapID string, ws []changelog.Warning) {
	if len(ws) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "warning: [%s] 改版履歴に %d 件の注意があります:\n", roadmapID, len(ws))
	for _, w := range ws {
		if w.NodeID != "" {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", w.NodeID, w.Message)
		} else {
			fmt.Fprintf(os.Stderr, "  - %s\n", w.Message)
		}
	}
}

func totalNodes(graphs []*graph.Graph) int {
	n := 0
	for _, g := range graphs {
		n += len(g.Nodes)
	}
	return n
}
