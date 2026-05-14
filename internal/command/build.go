package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/meta"
	"github.com/fuchigta/roadmapper/internal/render"
	"github.com/fuchigta/roadmapper/web"
)

func NewBuildCmd() *cobra.Command {
	var (
		configPath string
		outDir     string
		basePath   string
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "静的サイトを生成する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(configPath, outDir, basePath)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "roadmap.yml", "設定ファイルのパス")
	cmd.Flags().StringVarP(&outDir, "out", "o", "dist", "出力ディレクトリ")
	cmd.Flags().StringVar(&basePath, "base", "", "ベースパス (例: /my-repo/)")

	return cmd
}

func runBuild(configPath, outDir, basePath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg); err != nil {
		return err
	}

	// basePath: フラグ優先、なければ config
	if basePath == "" {
		basePath = cfg.Site.BasePath
	}

	// roadmap.yml と同じディレクトリを content/ のベースとする
	configDir := filepath.Dir(configPath)
	contentDir := filepath.Join(configDir, "content")

	// content/ の全 Markdown をロード
	docs, err := content.LoadDir(contentDir)
	if err != nil {
		return fmt.Errorf("content ディレクトリの読み込みに失敗: %w", err)
	}

	// 出力ディレクトリを作成
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("出力ディレクトリを作成できません: %w", err)
	}

	// 静的アセット (style.css / app.js) をコピー
	if err := copyStaticAssets(outDir); err != nil {
		return err
	}

	// content/ 配下の非 .md ファイルを dist/content/ にコピー
	assets, err := content.LoadAssets(contentDir, cfg.Site.ContentAssets.Exclude)
	if err != nil {
		return err
	}
	if err := copyContentAssets(outDir, assets); err != nil {
		return err
	}

	// assetBase: basePath が空のときはルートへの相対パス "../"
	assetBase := "../"
	if basePath != "" {
		assetBase = basePath
		if !strings.HasSuffix(assetBase, "/") {
			assetBase += "/"
		}
	}

	// ロードマップ → グラフ のマップ (index ページのカード進捗に使う)
	graphs := map[string]*graph.Graph{}

	// 各ロードマップを処理
	for i := range cfg.Roadmaps {
		rm := &cfg.Roadmaps[i]
		fmt.Printf("  処理中: %s\n", rm.ID)

		g, err := graph.Build(rm)
		if err != nil {
			return fmt.Errorf("ロードマップ %q のグラフ構築に失敗: %w", rm.ID, err)
		}
		graphs[rm.ID] = g

		lr, err := layout.Compute(g, cfg)
		if err != nil {
			return fmt.Errorf("ロードマップ %q のレイアウト計算に失敗: %w", rm.ID, err)
		}

		// ノード本文を Markdown → HTML / plaintext に変換
		nodeHTML, nodeText, hasMermaid, err := buildNodeHTML(g, docs, assetBase)
		if err != nil {
			return err
		}

		// ロードマップ用ディレクトリ
		rmDir := filepath.Join(outDir, rm.ID)
		if err := os.MkdirAll(rmDir, 0o755); err != nil {
			return fmt.Errorf("ディレクトリ作成失敗: %w", err)
		}

		pageHTML, err := render.RenderRoadmapPage(
			web.FS, cfg, rm, g, lr, nodeHTML, nodeText, basePath, assetBase, hasMermaid,
		)
		if err != nil {
			return fmt.Errorf("ロードマップページの生成に失敗: %w", err)
		}

		if err := os.WriteFile(filepath.Join(rmDir, "index.html"), []byte(pageHTML), 0o644); err != nil {
			return fmt.Errorf("index.html の書き込みに失敗: %w", err)
		}
	}

	// index.html 生成
	indexHTML, err := render.RenderIndexPage(web.FS, cfg, basePath, graphs)
	if err != nil {
		return fmt.Errorf("インデックスページの生成に失敗: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "index.html"), []byte(indexHTML), 0o644); err != nil {
		return fmt.Errorf("index.html の書き込みに失敗: %w", err)
	}

	// sitemap.xml 生成
	if sitemapXML, err := meta.RenderSitemap(cfg); err != nil {
		return fmt.Errorf("sitemap.xml の生成に失敗: %w", err)
	} else if sitemapXML != "" {
		if err := os.WriteFile(filepath.Join(outDir, "sitemap.xml"), []byte(sitemapXML), 0o644); err != nil {
			return fmt.Errorf("sitemap.xml の書き込みに失敗: %w", err)
		}
		fmt.Println("  sitemap.xml を生成しました")
	}

	// feed.rss 生成
	if rssXML, err := meta.RenderRSS(cfg, graphs); err != nil {
		return fmt.Errorf("feed.rss の生成に失敗: %w", err)
	} else if rssXML != "" {
		if err := os.WriteFile(filepath.Join(outDir, "feed.rss"), []byte(rssXML), 0o644); err != nil {
			return fmt.Errorf("feed.rss の書き込みに失敗: %w", err)
		}
		fmt.Println("  feed.rss を生成しました")
	}

	fmt.Printf("\n✓ %s に出力しました\n", outDir)
	return nil
}

// buildNodeHTML は各ノードの Markdown を HTML / plaintext に変換して map 2 つと mermaid 有無を返す。
func buildNodeHTML(g *graph.Graph, docs map[string]*content.Doc, assetBase string) (map[string]string, map[string]string, bool, error) {
	nodeHTML := map[string]string{}
	nodeText := map[string]string{}
	hasMermaid := false

	var unresolved []string
	for _, n := range g.Nodes {
		doc, ok := lookupDoc(docs, n.Node)
		if !ok {
			nodeHTML[n.ID] = ""
			nodeText[n.ID] = ""
			if n.Node.Content != "" {
				unresolved = append(unresolved, fmt.Sprintf("%s (content: %q)", n.ID, n.Node.Content))
			} else {
				unresolved = append(unresolved, n.ID)
			}
			continue
		}

		urlPrefix := assetBase + "content/"
		if doc.RelDir != "" {
			urlPrefix += doc.RelDir + "/"
		}
		html, err := render.RenderMarkdownWithBase(doc.Body, urlPrefix)
		if err != nil {
			return nil, nil, false, fmt.Errorf("ノード %q の Markdown 変換に失敗: %w", n.ID, err)
		}

		// links が content frontmatter にあれば config.Node の Links に追加 (content 優先)
		if len(doc.Frontmatter.Links) > 0 {
			links := make([]config.Link, len(doc.Frontmatter.Links))
			for i, l := range doc.Frontmatter.Links {
				links[i] = config.Link{Title: l.Title, URL: l.URL}
			}
			n.Node.Links = links
		}

		if render.HasMermaid(html) {
			hasMermaid = true
		}

		// リンク集を本文の後ろに追記
		if len(n.Node.Links) > 0 {
			html += render.RenderLinks(n.Node.Links)
		}

		nodeHTML[n.ID] = html

		// 全文検索用 plaintext: 本文テキスト + リンクタイトル
		plainBody := render.ExtractPlainText(doc.Body)
		linkTitles := make([]string, 0, len(n.Node.Links))
		for _, l := range n.Node.Links {
			if l.Title != "" {
				linkTitles = append(linkTitles, l.Title)
			}
		}
		linkText := strings.Join(linkTitles, " ")
		if linkText != "" {
			nodeText[n.ID] = strings.TrimSpace(plainBody + " " + linkText)
		} else {
			nodeText[n.ID] = plainBody
		}
	}

	if len(unresolved) > 0 {
		fmt.Fprintf(os.Stderr, "  warning: %d 個のノードに対応する content/*.md が見つかりません:\n", len(unresolved))
		for _, id := range unresolved {
			fmt.Fprintf(os.Stderr, "    - %s\n", id)
		}
	}

	return nodeHTML, nodeText, hasMermaid, nil
}

// lookupDoc はノードに対応する content.Doc を 3 段階で探す:
//  1. node.Content が指定されていればそのキーで引く (拡張子なし、スラッシュ区切り)
//  2. node.ID を完全キーとして引く (例: "frontend/html" → docs["frontend/html"])
//  3. node.ID を末尾名フォールバックキーとして引く (例: "html" → サブディレクトリ配下の docs["html"])
//
// ローダが (2) と (3) の両キーを登録しているため、実質的には docs[key] 一発で済む。
// node.Content 指定時はフォールバックを行わない (明示パスに従う)。
func lookupDoc(docs map[string]*content.Doc, n *config.Node) (*content.Doc, bool) {
	if n.Content != "" {
		doc, ok := docs[n.Content]
		return doc, ok
	}
	doc, ok := docs[n.ID]
	return doc, ok
}

// copyContentAssets は content/ 配下のアセットを outDir/content/ にコピーする。
func copyContentAssets(outDir string, assets []content.Asset) error {
	for _, a := range assets {
		dest := filepath.Join(outDir, "content", filepath.FromSlash(a.RelPath))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("アセット出力ディレクトリ作成失敗: %w", err)
		}
		if err := copyFile(a.SrcPath, dest); err != nil {
			return fmt.Errorf("アセット %s のコピーに失敗: %w", a.RelPath, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func copyStaticAssets(outDir string) error {
	entries := []string{"static/style.css", "static/app.js"}
	for _, entry := range entries {
		data, err := web.FS.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("アセット %s の読み込みに失敗: %w", entry, err)
		}
		dest := filepath.Join(outDir, filepath.Base(entry))
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return fmt.Errorf("アセット %s の書き込みに失敗: %w", entry, err)
		}
	}
	return nil
}
