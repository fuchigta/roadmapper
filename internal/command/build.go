package command

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/fuchigta/roadmapper/internal/changelog"
	"github.com/fuchigta/roadmapper/internal/config"
	"github.com/fuchigta/roadmapper/internal/content"
	"github.com/fuchigta/roadmapper/internal/graph"
	"github.com/fuchigta/roadmapper/internal/layout"
	"github.com/fuchigta/roadmapper/internal/meta"
	"github.com/fuchigta/roadmapper/internal/render"
	"github.com/fuchigta/roadmapper/internal/repo"
	"github.com/fuchigta/roadmapper/web"
)

// recentUpdateDays は SVG の「更新」バッジを表示する日数。
const recentUpdateDays = 30

func NewBuildCmd() *cobra.Command {
	var (
		configPath    string
		outDir        string
		basePath      string
		noAnalytics   bool
		includeDrafts bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "静的サイトを生成する",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(configPath, outDir, basePath, noAnalytics, includeDrafts)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "roadmap.yml", "設定ファイルのパス")
	cmd.Flags().StringVarP(&outDir, "out", "o", "dist", "出力ディレクトリ")
	cmd.Flags().StringVar(&basePath, "base", "", "ベースパス (例: /my-repo/)")
	cmd.Flags().BoolVar(&noAnalytics, "no-analytics", false, "アクセス解析タグを出力しない")
	cmd.Flags().BoolVar(&includeDrafts, "drafts", false, "下書きノードを通常ノードとして出力する (ステージング用)")

	return cmd
}

// runBuild はサイトを生成する。noAnalytics が true なら site.analytics を無効化する。
// includeDrafts が true なら下書きノードを通常ノードとして出力し「下書き」バッジを付ける (dev / build --drafts)。
// false なら下書きは非活性表示にし、本文・リンク・改版履歴を出力しない。
func runBuild(configPath, outDir, basePath string, noAnalytics, includeDrafts bool) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	if err := config.Validate(cfg); err != nil {
		return err
	}

	if noAnalytics {
		cfg.Site.Analytics.Provider = ""
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

	// 「この記事を編集」リンク用: configDir のリポジトリルートからの相対パス
	repoPrefix, err := repo.PathPrefix(configDir)
	if err != nil {
		return fmt.Errorf("リポジトリルートの検出に失敗: %w", err)
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
	// ロードマップ → 最新更新日 (index ページのカードに使う)
	latest := map[string]time.Time{}
	// ロードマップ → 改版履歴のマップ (RSS に使う)
	logs := map[string]*changelog.Log{}
	now := time.Now()
	// 下書きノード (ロードマップ ID → ノード ID 集合)。進捗の分母・ログ表示に使う
	draftsByRoadmap := map[string]map[string]bool{}

	// 各ロードマップを処理
	for i := range cfg.Roadmaps {
		rm := &cfg.Roadmaps[i]
		fmt.Printf("  処理中: %s\n", rm.ID)

		g, err := graph.Build(rm)
		if err != nil {
			return fmt.Errorf("ロードマップ %q のグラフ構築に失敗: %w", rm.ID, err)
		}
		graphs[rm.ID] = g

		// 下書きノード。公開ビルドでは本文・履歴・編集パスを出力しない (locked)
		drafts := resolveDrafts(g, docs)
		draftsByRoadmap[rm.ID] = drafts
		var locked map[string]bool
		if !includeDrafts {
			locked = drafts
		}

		// 改版履歴を集約 (公開ビルドでは下書きの frontmatter と changelog の下書き参照を除く)
		log, err := changelog.Build(changelog.StripNodes(rm.Changelog, locked), resolveNodeDocs(g, docs, locked))
		if err != nil {
			return fmt.Errorf("ロードマップ %q の改版履歴の構築に失敗: %w", rm.ID, err)
		}
		latest[rm.ID] = log.Latest
		logs[rm.ID] = log

		// 最近更新されたノード (SVG 更新バッジ用)
		recent := changelog.Recent(log.NodeUpdated, now, recentUpdateDays)
		badges := make(map[string]time.Time, len(recent))
		for id := range recent {
			badges[id] = log.NodeUpdated[id]
		}

		lr, err := layout.Compute(g, cfg)
		if err != nil {
			return fmt.Errorf("ロードマップ %q のレイアウト計算に失敗: %w", rm.ID, err)
		}

		// ノード本文を Markdown → HTML / plaintext に変換
		nodeHTML, nodeText, hasMermaid, err := buildNodeHTML(g, docs, assetBase, locked)
		if err != nil {
			return err
		}

		editPaths := buildEditPaths(g, docs, repoPrefix, locked)

		// ロードマップ用ディレクトリ
		rmDir := filepath.Join(outDir, rm.ID)
		if err := os.MkdirAll(rmDir, 0o755); err != nil {
			return fmt.Errorf("ディレクトリ作成失敗: %w", err)
		}

		pageHTML, err := render.RenderRoadmapPage(
			web.FS, cfg, rm, g, lr, nodeHTML, nodeText, basePath, assetBase, hasMermaid, log, badges,
			editPaths,
			render.Drafts{IDs: drafts, Preview: includeDrafts},
		)
		if err != nil {
			return fmt.Errorf("ロードマップページの生成に失敗: %w", err)
		}

		if err := os.WriteFile(filepath.Join(rmDir, "index.html"), []byte(pageHTML), 0o644); err != nil {
			return fmt.Errorf("index.html の書き込みに失敗: %w", err)
		}
	}

	// index.html 生成
	indexHTML, err := render.RenderIndexPage(web.FS, cfg, basePath, graphs, latest, lockedByRoadmap(draftsByRoadmap, includeDrafts))
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
	if rssXML, err := meta.RenderRSS(cfg, graphs, logs); err != nil {
		return fmt.Errorf("feed.rss の生成に失敗: %w", err)
	} else if rssXML != "" {
		if err := os.WriteFile(filepath.Join(outDir, "feed.rss"), []byte(rssXML), 0o644); err != nil {
			return fmt.Errorf("feed.rss の書き込みに失敗: %w", err)
		}
		fmt.Println("  feed.rss を生成しました")
	}

	printDrafts(cfg, draftsByRoadmap, includeDrafts)

	fmt.Printf("\n✓ %s に出力しました\n", outDir)
	return nil
}

// buildNodeHTML は各ノードの Markdown を HTML / plaintext に変換して map 2 つと mermaid 有無を返す。
// locked のノード (公開ビルドの下書き) は本文を読まず空文字列を返す。
func buildNodeHTML(g *graph.Graph, docs map[string]*content.Doc, assetBase string, locked map[string]bool) (map[string]string, map[string]string, bool, error) {
	nodeHTML := map[string]string{}
	nodeText := map[string]string{}
	hasMermaid := false

	var unresolved []string
	for _, n := range g.Nodes {
		if locked[n.ID] {
			nodeHTML[n.ID], nodeText[n.ID] = "", ""
			continue
		}
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

// buildEditPaths は各ノードの記事ファイルのリポジトリルート相対パスを返す。
// 記事が存在しないノードは、作成先となる推定パス (<prefix>content/<content or id>.md) を返す。
// locked のノード (公開ビルドの下書き) は記事の場所を出さない。
func buildEditPaths(g *graph.Graph, docs map[string]*content.Doc, prefix string, locked map[string]bool) map[string]string {
	out := make(map[string]string, len(g.Nodes))
	for _, n := range g.Nodes {
		if locked[n.ID] {
			continue
		}
		if doc, ok := lookupDoc(docs, n.Node); ok && doc.RelPath != "" {
			out[n.ID] = prefix + "content/" + doc.RelPath
			continue
		}
		key := n.ID
		if n.Node.Content != "" {
			key = n.Node.Content
		}
		out[n.ID] = prefix + "content/" + key + ".md"
	}
	return out
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

// resolveDrafts は下書きノードの ID 集合を返す。
// roadmap.yml のノードの draft: true、または対応する記事の frontmatter の draft: true のどちらかで下書きになる。
func resolveDrafts(g *graph.Graph, docs map[string]*content.Doc) map[string]bool {
	out := map[string]bool{}
	for _, n := range g.Nodes {
		if n.Node.Draft {
			out[n.ID] = true
			continue
		}
		if doc, ok := lookupDoc(docs, n.Node); ok && doc.Frontmatter.Draft {
			out[n.ID] = true
		}
	}
	return out
}

// lockedByRoadmap は index ページの進捗から除くノードを返す。プレビュー (includeDrafts) では除外しない。
func lockedByRoadmap(drafts map[string]map[string]bool, includeDrafts bool) map[string]map[string]bool {
	if includeDrafts {
		return nil
	}
	return drafts
}

// printDrafts は下書きノードの件数と ID を表示する。
func printDrafts(cfg *config.Config, drafts map[string]map[string]bool, includeDrafts bool) {
	var lines []string
	for _, rm := range cfg.Roadmaps {
		ids := make([]string, 0, len(drafts[rm.ID]))
		for id := range drafts[rm.ID] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			lines = append(lines, fmt.Sprintf("    - [%s] %s", rm.ID, id))
		}
	}
	if len(lines) == 0 {
		return
	}
	mode := "非公開"
	if includeDrafts {
		mode = "プレビュー表示"
	}
	fmt.Printf("  下書き %d 件 (%s):\n%s\n", len(lines), mode, strings.Join(lines, "\n"))
}
