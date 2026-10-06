# CLAUDE.md — roadmapper 開発ガイド

## プロジェクト概要

Go 製の学習ロードマップ静的サイトジェネレータ CLI。`roadmap.yml` + `content/*.md` から
GitHub Pages / GitLab Pages 対応の静的サイトを生成する。**外部ランタイム依存ゼロ** が最重要設計原則。

## Git フック (lefthook)

```bash
# 初回セットアップ (クローン後に実行)
lefthook install

# 手動実行
lefthook run pre-commit   # gofmt / go vet / staticcheck
lefthook run pre-push     # go test -race ./...
```

- `pre-commit`: gofmt・go vet・staticcheck を並列実行 (staged Go ファイル対象)
- `commit-msg`: Conventional Commits 形式を検証 (`.lefthook/commit-msg/conventional.sh`)
- `pre-push`: `go test -race ./...` を全パッケージに実行

## 必須コマンド

```bash
# ビルド
go build ./...

# テスト
go test ./...

# 単一パッケージテスト
go test ./internal/render/...

# linter (golangci-lint がある場合)
golangci-lint run

# 動作確認
go run ./cmd/roadmapper init demo --template frontend-beginner
go run ./cmd/roadmapper build -c demo/roadmap.yml -o demo/dist
go run ./cmd/roadmapper validate -c demo/roadmap.yml   # --strict で警告もエラー扱い
go run ./cmd/roadmapper dev -c demo/roadmap.yml   # Ctrl+C で終了
```

## ディレクトリ構造

```
cmd/roadmapper/main.go          # CLIエントリ (cobra)
internal/
  command/                      # CLI サブコマンド実装
    build.go                    # roadmapper build
    changelog.go                # 改版履歴用のノード→Doc 解決ヘルパ (サブコマンドではない)
    dev.go                      # roadmapper dev (fsnotify + SSE livereload)
    deploy.go                   # roadmapper deploy --target github|gitlab
    init.go                     # roadmapper init --template minimal|frontend-beginner|backend-beginner|devops|blank
    validate.go                 # roadmapper validate [--strict]
  changelog/                    # 改版履歴 (roadmap.yml + frontmatter の集約・検証・Recent 判定)
    changelog.go                # Build / Check / Recent (純粋関数)
  config/                       # roadmap.yml パーサ + バリデーション
    schema.go                   # Config / Site / Roadmap / Node / Link 構造体
    loader.go                   # Load(path) → *Config (applyDefaults で既定値補完)
    validate.go                 # Validate(*Config) error (site / node / panel / progressSync / changelog)
  content/                      # content/**/*.md ローダ (サブディレクトリ対応)
    loader.go                   # LoadDir(dir) → map[string]*Doc (再帰スキャン、相対パス優先・ファイル名末尾フォールバック)
                                # LoadAssets: 非 .md ファイルの収集 (contentAssets.exclude 適用)
                                # frontmatter: title / links / updated / changes
    glob.go                     # MatchGlob / MatchAny (`*` 単一セグメント・`**` 複数セグメント)
  graph/                        # ノード/エッジ DAG モデル
    graph.go                    # Build(*Roadmap) → *Graph (cycle detection)
  layout/                       # Goja + dagre.js でレイアウト計算
    layout.go                   # Compute(*Graph, *Config) → *Result
    vendor/dagre.min.js         # //go:embed で同梱
  meta/                         # sitemap / RSS (OGP の og:url 用 URL 組み立て含む)
    base.go                     # SiteBase(siteURL, basePath) → 末尾スラッシュ付きベース URL
    sitemap.go                  # RenderSitemap(*Config) → XML string (siteUrl 空なら "")
    rss.go                      # RenderRSS(*Config, graphs, logs) → XML string (siteUrl 空なら "")
  render/                       # HTML / SVG / Markdown レンダリング
    svg.go                      # RenderSVG / RenderSVGWithBadges(…, updated) → SVG string
    html.go                     # RenderRoadmapPage / RenderIndexPage → HTML string
    changelog.go                # 改版履歴 (ヘッダー / パネル) の HTML 断片
    markdown.go                 # RenderMarkdown / RenderMarkdownWithBase(body, urlPrefix) → HTML string (goldmark + chroma)
    links.go                    # RenderLinks([]Link) → HTML fragment
    theme.go                    # DeriveColors(hex) → {Base, Light}
  server/                       # dev サーバ
    server.go                   # HTTP server + SSE + livereload script injection
  templates/                    # init コマンド用スケルトン
    embed.go                    # //go:embed all:data
    templates_test.go           # 全テンプレートが strict 検証を通ることを確認
    data/minimal/               # 最小サンプル
    data/frontend-beginner/     # 現実的テンプレート (改版履歴の例入り)
    data/backend-beginner/      # バックエンド学習テンプレート
    data/devops/                # DevOps / インフラ学習テンプレート
    data/blank/                 # roadmap.yml + content/start.md のみ
web/                            # ビルド時埋め込みアセット
  embed.go                      # //go:embed templates static
  sync_http_test.go             # progressSync の HTTP 部分を Goja で検証 (app.js の TEST:HTTP_BEGIN〜END を抽出)
  sync_merge_test.go            # マージ規則 (TEST:MERGE_BEGIN〜END) の検証
  templates/index.html          # インデックスページ
  templates/roadmap.html        # ロードマップページ
  static/style.css              # CSS variables ベースのテーマ
  static/app.js                 # 進捗トラッキング / サイドパネル / 検索 / シェア / progressSync / テーマ切替
docs/                           # roadmapper 自身で作った使い方ガイドサイト (docs/roadmap.yml + docs/content/*.md)
```

`web/static/app.js` の `// TEST:*_BEGIN` / `// TEST:*_END` コメントは Go テストが該当範囲を抜き出すための目印なので削除しない。

## 重要な実装規則

### embed FS のパス区切り
- `//go:embed` の FS は常に `/` 区切り。Windows でも `path.Join`（`filepath.Join` ではない）を使う
- `internal/templates/embed.go` 参照

### URL 結合
- `siteURL + basePath` を結合するとき必ず trailing slash を確保する
  ```go
  if !strings.HasSuffix(basePath, "/") { basePath += "/" }
  base := strings.TrimRight(siteURL, "/") + basePath
  ```

### basePath vs assetBase
- `basePath`: ページ間リンク URL (例: `/my-repo/`)
- `assetBase`: CSS/JS アセットへの相対パス。`basePath` が空なら `"../"` (サブディレクトリから root へ戻る)

### Mermaid パススルー
- goldmark の AST レンダラーは登録しない。Markdown → HTML 後に正規表現で後処理する
- `render/markdown.go` の `mermaidBlockRe` 参照

### チェックリストの `disabled` 属性
- goldmark GFM タスクリストは `<input disabled="">` を生成する
- 正規表現 ` disabled=""` を削除して操作可能にしている

### フロントエンドの進捗データ
- localStorage key: `roadmapper:progress`
- 構造: `{ [roadmapId]: { [nodeId]: { state, tasks[] } } }`
- `state`: `none` / `in-progress` / `done` / `skipped`
- チェックリストの初期チェック状態 (`- [x]`) は反映されない (`app.js` が localStorage の `tasks` で上書きする)

### 進捗バックエンド同期 (`progressSync`)
- `roadmap.yml` の `site.progressSync.enabled: true` + `endpoint: <BASE_URL>` で有効化
- デバイス匿名 UUID を `roadmapper:deviceId` キーで localStorage に保存
- `endpoint` は `http://` / `https://` 始まり必須 (`config.Validate` で検証)
- `GET {endpoint}/{deviceId}/{roadmapId}` で起動時にリモート進捗を取得し、ローカルとマージ (404 は空データ扱い、その他の失敗は無視)
- `PUT {endpoint}/{deviceId}/{roadmapId}` で状態変更 800ms 後にデバウンス送信
- マージ規則: ノードごとに「進んでいる方」を採用。`STATE_RANK`: none(0) < in-progress(1) < skipped/done(2)、同ランクは done 優先
- オフライン時は `roadmapper:sync-dirty:{roadmapId}` フラグを立て、`online` イベントまたは次回ロード時に再送
- シェアビュー (`?p=...`) では同期を行わない

### 改版履歴 (changelog) / panel / contentAssets
- 改版履歴は `roadmap.yml` の `roadmaps[].changelog` と content frontmatter の `updated` / `changes` を `changelog.Build` で集約する。日付は `YYYY-MM-DD`。形式不正は `config.Validate` / `changelog.Build` でエラー、整合性の問題 (未来日付・`updated` が履歴より古い等) は `changelog.Check` の警告 (`validate --strict` で失敗)
- SVG の「更新」バッジは `changelog.Recent` で最終更新から 30 日以内のノード (`build.go` の `recentUpdateDays`)
- `site.panel` (width / minWidth / maxWidth) の既定値補完は `applyDefaults`、整合性検証は `validatePanel`
- `site.contentAssets.exclude` は `content.LoadAssets` で `content/` からの相対パスに glob 適用。`content/` の非 `.md` ファイルは `dist/content/` にコピーされ、Markdown 内の相対 URL は `RenderMarkdownWithBase` で書き換える

## 禁止事項

- **外部バイナリ依存の追加禁止** — Node.js, graphviz, Python, etc. はインストール不要のまま保つ
- **フロントエンドフレームワーク追加禁止** — `web/static/app.js` は素の JS のまま維持する (目標 32KB 以内。現状 約 30KB)
- **`filepath.Join` を embed FS パスに使用禁止** — `path.Join` を使うこと
- **`web/` 以下のファイルをビルド外から直接コピーしない** — `web.FS` 経由でアクセスする

### content/ とノード ID の対応規則

`content/` は `filepath.WalkDir` で再帰的にスキャンし、map キーを 2 種類登録する:

- **相対パスキー** (常に登録): `content/frontend/html.md` → `"frontend/html"`
- **末尾名フォールバックキー** (一意な場合のみ): `"html"` → 同じ Doc を指す

優先規則: ルート直下ファイル (`content/html.md` → `"html"`) が優先。複数サブディレクトリに
同名ファイルがある場合、フォールバックキーは**曖昧として登録されない** (`docs["html"]` が nil
になる)。その場合は相対パスキー (`"frontend/html"`) で参照すること。

## テスト方針

- 各 `internal/` パッケージにユニットテストを置く
- ゴールデンファイルテストは `testdata/` ディレクトリに配置
- `go test ./...` がすべて通ること
- `internal/command/` にはテストファイルなし (統合テストは手動確認)
- `internal/templates/templates_test.go` が全テンプレートの strict 検証を行うため、テンプレート追加・変更時は `go test ./internal/templates/...` を通すこと
- `web/` の JS ロジックは Goja 上で Go テストから検証する (`sync_http_test.go` / `sync_merge_test.go`)

## 依存ライブラリ

| ライブラリ | 用途 |
|---|---|
| `github.com/spf13/cobra` | CLI フレームワーク |
| `github.com/dop251/goja` | Pure Go JS エンジン (dagre.js 実行) |
| `github.com/yuin/goldmark` | Markdown → HTML |
| `github.com/yuin/goldmark-highlighting/v2` | goldmark と chroma の連携 |
| `github.com/alecthomas/chroma/v2` | シンタックスハイライト |
| `github.com/fsnotify/fsnotify` | ファイル監視 (dev コマンド) |
| `gopkg.in/yaml.v3` | YAML パース |
