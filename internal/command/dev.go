package command

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/server"
)

func NewDevCmd() *cobra.Command {
	var (
		configPath string
		outDir     string
		port       int
	)

	cmd := &cobra.Command{
		Use:   "dev",
		Short: "開発サーバを起動してファイル変更を監視する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDev(configPath, outDir, port)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "roadmap.yml", "設定ファイルのパス")
	cmd.Flags().StringVarP(&outDir, "out", "o", "dist", "出力ディレクトリ")
	cmd.Flags().IntVarP(&port, "port", "p", 4321, "開発サーバのポート番号")

	return cmd
}

func runDev(configPath, outDir string, port int) error {
	// 初回ビルド
	fmt.Println("初回ビルド中...")
	if err := runBuild(configPath, outDir, ""); err != nil {
		return fmt.Errorf("初回ビルド失敗: %w", err)
	}

	srv := server.New(outDir)

	// ファイル監視
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("ウォッチャー作成失敗: %w", err)
	}
	defer watcher.Close()

	// roadmap.yml と content/ を監視
	configDir := filepath.Dir(configPath)
	watchTargets := []string{configPath, filepath.Join(configDir, "content")}
	for _, t := range watchTargets {
		if _, err := os.Stat(t); err == nil {
			if err := watcher.Add(t); err != nil {
				log.Printf("監視追加失敗 %s: %v", t, err)
			}
		}
	}

	// デバウンス付きリビルドゴルーチン
	go func() {
		var timer *time.Timer
		debounce := 300 * time.Millisecond

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) {
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(debounce, func() {
						fmt.Printf("\nファイル変更検出: %s\nリビルド中...\n", event.Name)
						if err := runBuild(configPath, outDir, ""); err != nil {
							log.Printf("ビルドエラー: %v", err)
						} else {
							srv.Notify()
							fmt.Println("リビルド完了、ブラウザを更新します")
						}
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("ウォッチャーエラー: %v", err)
			}
		}
	}()

	fmt.Printf("\n✓ 開発サーバを起動しました\n")
	fmt.Printf("  URL: http://localhost:%d\n", port)
	fmt.Printf("  監視対象: %s, %s/content/\n", configPath, configDir)
	fmt.Printf("  (Ctrl+C で終了)\n\n")

	return srv.Start(port)
}
