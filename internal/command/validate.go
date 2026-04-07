package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/graph"
)

func NewValidateCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "roadmap.yml と content/ の整合性を検証する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(configPath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "roadmap.yml",
		"設定ファイルのパス")

	return cmd
}

func runValidate(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return fmt.Errorf("バリデーションに失敗しました")
	}

	// グラフ構造の検証 (循環参照など)
	for _, rm := range cfg.Roadmaps {
		rm := rm
		if _, err := graph.Build(&rm); err != nil {
			return fmt.Errorf("ロードマップ %q: %w", rm.ID, err)
		}
	}

	fmt.Printf("✓ %s の検証が完了しました (%d ロードマップ)\n",
		configPath, len(cfg.Roadmaps))
	return nil
}
