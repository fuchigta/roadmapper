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
	cfg, err := loadBuildConfig(configPath, noAnalytics)
	if err != nil {
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

	// 「この記事を編集」リンク用: configDir のリポジトリルートからの相対パス
	repoPrefix, err := repo.PathPrefix(configDir)
	if err != nil {
		return fmt.Errorf("リポジトリルートの検出に失敗: %w", err)
	}

	if err := prepareOutDir(outDir, contentDir, cfg.Site.ContentAssets.Exclude); err != nil {
		return err
	}

	env := &buildEnv{
		cfg:           cfg,
		docs:          docs,
		outDir:        outDir,
		basePath:      basePath,
		assetBase:     resolveAssetBase(basePath),
		repoPrefix:    repoPrefix,
		includeDrafts: includeDrafts,
		now:           time.Now(),
	}
	res := newBuildResults()

	// 各ロードマップを処理
	for i := range cfg.Roadmaps {
		if err := env.buildRoadmap(&cfg.Roadmaps[i], res); err != nil {
			return err
		}
	}

	if err := env.writeSiteFiles(res); err != nil {
		return err
	}

	printDrafts(cfg, res.drafts, includeDrafts)

	fmt.Printf("\n✓ %s に出力しました\n", outDir)
	return nil
}

// loadBuildConfig は roadmap.yml を読み込んで検証する。noAnalytics が true なら site.analytics を無効化する。
func loadBuildConfig(configPath string, noAnalytics bool) (*config.Config, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}
	if noAnalytics {
		cfg.Site.Analytics.Provider = ""
	}
	return cfg, nil
}

// prepareOutDir は出力ディレクトリを作成し、静的アセット (style.css / app.js) と
// content/ 配下の非 .md ファイル (dist/content/ へ) をコピーする。
func prepareOutDir(outDir, contentDir string, exclude []string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("出力ディレクトリを作成できません: %w", err)
	}
	if err := copyStaticAssets(outDir); err != nil {
		return err
	}
	assets, err := content.LoadAssets(contentDir, exclude)
	if err != nil {
		return err
	}
	return copyContentAssets(outDir, assets)
}

// resolveAssetBase は CSS/JS アセットの参照ベースを返す。
// basePath が空のときはルートへの相対パス "../"、それ以外は末尾スラッシュ付きの basePath。
func resolveAssetBase(basePath string) string {
	if basePath == "" {
		return "../"
	}
	if !strings.HasSuffix(basePath, "/") {
		return basePath + "/"
	}
	return basePath
}

// buildEnv はビルド全体で共有する入力。
type buildEnv struct {
	cfg           *config.Config
	docs          map[string]*content.Doc
	outDir        string
	basePath      string
	assetBase     string
	repoPrefix    string
	includeDrafts bool
	now           time.Time
}

// buildResults はロードマップごとの処理結果。index / RSS / 下書き表示に使う。
type buildResults struct {
	// graphs はロードマップ → グラフ (index ページのカード進捗に使う)
	graphs map[string]*graph.Graph
	// latest はロードマップ → 最新更新日 (index ページのカードに使う)
	latest map[string]time.Time
	// logs はロードマップ → 改版履歴 (RSS に使う)
	logs map[string]*changelog.Log
	// drafts は下書きノード (ロードマップ ID → ノード ID 集合)。進捗の分母・ログ表示に使う
	drafts map[string]map[string]bool
}

func newBuildResults() *buildResults {
	return &buildResults{
		graphs: map[string]*graph.Graph{},
		latest: map[string]time.Time{},
		logs:   map[string]*changelog.Log{},
		drafts: map[string]map[string]bool{},
	}
}

// buildRoadmap は 1 つのロードマップのページを outDir/<id>/index.html に出力し、結果を res に記録する。
func (env *buildEnv) buildRoadmap(rm *config.Roadmap, res *buildResults) error {
	fmt.Printf("  処理中: %s\n", rm.ID)

	g, err := graph.Build(rm)
	if err != nil {
		return fmt.Errorf("ロードマップ %q のグラフ構築に失敗: %w", rm.ID, err)
	}
	res.graphs[rm.ID] = g

	// 下書きノード。公開ビルドでは本文・履歴・編集パスを出力しない (locked)
	drafts := resolveDrafts(g, env.docs)
	res.drafts[rm.ID] = drafts
	var locked map[string]bool
	if !env.includeDrafts {
		locked = drafts
	}

	// 改版履歴を集約 (公開ビルドでは下書きの frontmatter と changelog の下書き参照を除く)
	log, err := changelog.Build(changelog.StripNodes(rm.Changelog, locked), resolveNodeDocs(g, env.docs, locked))
	if err != nil {
		return fmt.Errorf("ロードマップ %q の改版履歴の構築に失敗: %w", rm.ID, err)
	}
	res.latest[rm.ID] = log.Latest
	res.logs[rm.ID] = log

	lr, err := layout.Compute(g, env.cfg)
	if err != nil {
		return fmt.Errorf("ロードマップ %q のレイアウト計算に失敗: %w", rm.ID, err)
	}

	// ノード本文を Markdown → HTML / plaintext に変換
	nodeHTML, nodeText, hasMermaid, err := buildNodeHTML(g, env.docs, env.assetBase, locked)
	if err != nil {
		return err
	}

	editPaths := buildEditPaths(g, env.docs, env.repoPrefix, locked)

	// ロードマップ用ディレクトリ
	rmDir := filepath.Join(env.outDir, rm.ID)
	if err := os.MkdirAll(rmDir, 0o755); err != nil {
		return fmt.Errorf("ディレクトリ作成失敗: %w", err)
	}

	pageHTML, err := render.RenderRoadmapPage(
		web.FS, env.cfg, rm, g, lr, nodeHTML, nodeText, env.basePath, env.assetBase, hasMermaid, log,
		recentBadges(log, env.now),
		editPaths,
		render.Drafts{IDs: drafts, Preview: env.includeDrafts},
	)
	if err != nil {
		return fmt.Errorf("ロードマップページの生成に失敗: %w", err)
	}

	if err := os.WriteFile(filepath.Join(rmDir, "index.html"), []byte(pageHTML), 0o644); err != nil {
		return fmt.Errorf("index.html の書き込みに失敗: %w", err)
	}
	return nil
}

// recentBadges は最近更新されたノード (SVG 更新バッジ用) → 最終更新日を返す。
func recentBadges(log *changelog.Log, now time.Time) map[string]time.Time {
	recent := changelog.Recent(log.NodeUpdated, now, recentUpdateDays)
	badges := make(map[string]time.Time, len(recent))
	for id := range recent {
		badges[id] = log.NodeUpdated[id]
	}
	return badges
}

// writeSiteFiles はサイト全体のファイル (index.html / sitemap.xml / feed.rss) を出力する。
func (env *buildEnv) writeSiteFiles(res *buildResults) error {
	indexHTML, err := render.RenderIndexPage(web.FS, env.cfg, env.basePath, res.graphs, res.latest, lockedByRoadmap(res.drafts, env.includeDrafts))
	if err != nil {
		return fmt.Errorf("インデックスページの生成に失敗: %w", err)
	}
	if err := os.WriteFile(filepath.Join(env.outDir, "index.html"), []byte(indexHTML), 0o644); err != nil {
		return fmt.Errorf("index.html の書き込みに失敗: %w", err)
	}

	if err := writeOptionalFile(env.outDir, "sitemap.xml", func() (string, error) {
		return meta.RenderSitemap(env.cfg)
	}); err != nil {
		return err
	}
	return writeOptionalFile(env.outDir, "feed.rss", func() (string, error) {
		return meta.RenderRSS(env.cfg, res.graphs, res.logs)
	})
}

// writeOptionalFile は generate の結果を outDir/name に書き出す。結果が空 (siteUrl 未設定など) なら何もしない。
func writeOptionalFile(outDir, name string, generate func() (string, error)) error {
	data, err := generate()
	if err != nil {
		return fmt.Errorf("%s の生成に失敗: %w", name, err)
	}
	if data == "" {
		return nil
	}
	if err := os.WriteFile(filepath.Join(outDir, name), []byte(data), 0o644); err != nil {
		return fmt.Errorf("%s の書き込みに失敗: %w", name, err)
	}
	fmt.Printf("  %s を生成しました\n", name)
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
			nodeHTML[n.ID], nodeText[n.ID] = "", ""
			unresolved = append(unresolved, unresolvedLabel(n.Node))
			continue
		}

		html, mermaid, err := renderNodeDoc(n.Node, doc, assetBase)
		if err != nil {
			return nil, nil, false, fmt.Errorf("ノード %q の Markdown 変換に失敗: %w", n.ID, err)
		}
		hasMermaid = hasMermaid || mermaid
		nodeHTML[n.ID] = html
		nodeText[n.ID] = nodePlainText(doc.Body, n.Node.Links)
	}

	warnUnresolvedDocs(unresolved)

	return nodeHTML, nodeText, hasMermaid, nil
}

// unresolvedLabel は記事が見つからないノードの警告表示用ラベルを返す。
func unresolvedLabel(n *config.Node) string {
	if n.Content != "" {
		return fmt.Sprintf("%s (content: %q)", n.ID, n.Content)
	}
	return n.ID
}

// renderNodeDoc はノードの記事を HTML に変換し、mermaid を含むかを返す。
// 記事 frontmatter に links があれば n.Links を置き換え (content 優先)、リンク集を本文の後ろに追記する。
func renderNodeDoc(n *config.Node, doc *content.Doc, assetBase string) (string, bool, error) {
	urlPrefix := assetBase + "content/"
	if doc.RelDir != "" {
		urlPrefix += doc.RelDir + "/"
	}
	html, err := render.RenderMarkdownWithBase(doc.Body, urlPrefix)
	if err != nil {
		return "", false, err
	}

	applyFrontmatterLinks(n, doc.Frontmatter.Links)
	hasMermaid := render.HasMermaid(html)

	// リンク集を本文の後ろに追記
	if len(n.Links) > 0 {
		html += render.RenderLinks(n.Links)
	}
	return html, hasMermaid, nil
}

// applyFrontmatterLinks は content frontmatter の links があれば config.Node の Links を置き換える (content 優先)。
func applyFrontmatterLinks(n *config.Node, fmLinks []content.Link) {
	if len(fmLinks) == 0 {
		return
	}
	links := make([]config.Link, len(fmLinks))
	for i, l := range fmLinks {
		links[i] = config.Link{Title: l.Title, URL: l.URL}
	}
	n.Links = links
}

// nodePlainText は全文検索用 plaintext (本文テキスト + 空でないリンクタイトル) を返す。
func nodePlainText(body string, links []config.Link) string {
	plainBody := render.ExtractPlainText(body)
	linkTitles := make([]string, 0, len(links))
	for _, l := range links {
		if l.Title != "" {
			linkTitles = append(linkTitles, l.Title)
		}
	}
	linkText := strings.Join(linkTitles, " ")
	if linkText == "" {
		return plainBody
	}
	return strings.TrimSpace(plainBody + " " + linkText)
}

// warnUnresolvedDocs は記事が見つからないノードを stderr に警告する。
func warnUnresolvedDocs(unresolved []string) {
	if len(unresolved) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "  warning: %d 個のノードに対応する content/*.md が見つかりません:\n", len(unresolved))
	for _, id := range unresolved {
		fmt.Fprintf(os.Stderr, "    - %s\n", id)
	}
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
