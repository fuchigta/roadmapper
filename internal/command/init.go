package command

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/templates"
)

const availableTemplates = "minimal | frontend-beginner | backend-beginner | devops | blank"

func NewInitCmd() *cobra.Command {
	var templateName string

	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "ロードマッププロジェクトを初期化する",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}
			return runInit(dir, templateName)
		},
	}

	cmd.Flags().StringVarP(&templateName, "template", "t", "minimal",
		"使用するテンプレート ("+availableTemplates+")")

	return cmd
}

func runInit(dir, templateName string) error {
	srcDir := path.Join("data", templateName)

	// テンプレートが存在するか確認
	if _, err := templates.FS.Open(srcDir); err != nil {
		return fmt.Errorf("テンプレート %q が見つかりません (%s)", templateName, availableTemplates)
	}

	// 出力先ディレクトリを作成
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ディレクトリを作成できません: %w", err)
	}

	// テンプレートをコピー
	err := fs.WalkDir(templates.FS, srcDir, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// srcDir 自体はスキップ
		relPath := strings.TrimPrefix(fpath, srcDir)
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" {
			return nil
		}

		// embed FS は常に / 区切り。OS のパス区切りに変換する。
		destPath := filepath.Join(dir, filepath.FromSlash(relPath))

		if d.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		data, err := templates.FS.ReadFile(fpath)
		if err != nil {
			return err
		}

		// 既にファイルが存在する場合はスキップ
		if _, err := os.Stat(destPath); err == nil {
			fmt.Printf("  スキップ (既存): %s\n", destPath)
			return nil
		}

		if err := os.WriteFile(destPath, data, 0o644); err != nil {
			return fmt.Errorf("%s の書き込みに失敗: %w", destPath, err)
		}
		fmt.Printf("  作成: %s\n", destPath)
		return nil
	})
	if err != nil {
		return fmt.Errorf("テンプレートのコピーに失敗しました: %w", err)
	}

	fmt.Printf("\n✓ %q を初期化しました (テンプレート: %s)\n", dir, templateName)
	fmt.Println("\n次のステップ:")
	fmt.Printf("  cd %s\n", dir)
	fmt.Println("  roadmapper validate   # 設定ファイルの検証")
	fmt.Println("  roadmapper build      # 静的サイトを生成")
	return nil
}
