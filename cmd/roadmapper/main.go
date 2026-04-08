package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/command"
)

// version はビルド時に -ldflags で注入される。
// goreleaser が "-X main.version=vX.Y.Z" を渡す。
var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "roadmapper",
	Version: version,
	Short:   "学習ロードマップ静的サイトジェネレータ",
	Long: `roadmapper は roadmap.yml と content/*.md から
静的サイト (GitHub Pages / GitLab Pages 対応) を生成する CLI ツールです。`,
}

func main() {
	rootCmd.AddCommand(
		command.NewInitCmd(),
		command.NewValidateCmd(),
		command.NewBuildCmd(),
		command.NewDevCmd(),
		command.NewDeployCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
