# CLAUDE.md — roadmapper 開発ガイド

## プロジェクト概要

Go 製の学習ロードマップ静的サイトジェネレータ CLI。`roadmap.yml` + `content/*.md` から
GitHub Pages / GitLab Pages 対応の静的サイトを生成する。**外部ランタイム依存ゼロ** が最重要設計原則。

## Git フック (lefthook) と spotter

```bash
# 初回セットアップ (クローン後に実行)
mise install              # mise.toml の spotter を導入 (mise を使わない場合は spotter の Releases から取得)
lefthook install

# 手動実行
lefthook run pre-commit   # gofmt / go vet / staticcheck
lefthook run pre-push     # go test -race ./... + spotter の全検査
spotter check --range origin/master..HEAD   # push 前の範囲を手元で検査
```

- `pre-commit`: gofmt・go vet・staticcheck を並列実行 (staged Go ファイル対象)
- `commit-msg`: Conventional Commits 形式を検証 (`spotter check commit-subject`)
- `pre-push`: `go test -race ./...` と `spotter check --pre-push` (CI の `test.yml` と同じ検査)

### ドキュメントと実装の一致 (`.spotter.yml`)

[spotter](https://github.com/fuchigta/spotter) で次を検査する。設定は `.spotter.yml`、バージョンは `mise.toml` と `.github/workflows/test.yml` で揃える。

- `doc-sync`: CLI フラグ・設定スキーマ・deploy の CI 雛形・解析イベント・ファイル追加削除・依存を変更したら、対応する README / ガイド / CLAUDE.md も同じ push 範囲で変更されていること (コミットを分けるのは可)
- `consistency`: フラグ一覧・テンプレート一覧・設定キー・解析イベント名が実装とドキュメントで一致すること、使い方ガイドの記事ファイル (`docs/content/*.md`) とノード ID が一対一であること
- `doc-paths` / `doc-links`: ドキュメント中のパス参照・リンク先が実在すること

違反したらドキュメントを直すのが原則。ドキュメントに影響しない変更なら、コミットメッセージ末尾のトレーラで理由付きで免除する (例: `Doc-Sync: skip[README.md] 内部リファクタでフラグは不変`)。
検査を追加・変更したら `spotter config lint` で死んだ設定がないか確認し、`spotter config explain <検査名>` で抽出結果 (consistency の要素・doc-paths の候補など) が意図どおりか確かめる。

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
    build.go                    # roadmapper build [--drafts] (resolveDrafts: 下書きノード判定)
    changelog.go                # 改版履歴用のノード→Doc 解決ヘルパ (サブコマンドではない)
    dev.go                      # roadmapper dev (fsnotify + SSE livereload)
    deploy.go                   # roadmapper deploy --target github|gitlab [--branch]
    init.go                     # roadmapper init --template minimal|frontend-beginner|backend-beginner|devops|blank
    validate.go                 # roadmapper validate [--strict]
  changelog/                    # 改版履歴 (roadmap.yml + frontmatter の集約・検証・Recent 判定)
    changelog.go                # Build / Check / Recent (純粋関数)
                                # StripNodes / CheckDrafts: 下書きノードの参照除去・検証
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
  repo/                         # .git 探索 (リポジトリルート検出 / 現在ブランチ読み取り、外部コマンド不使用)
    repo.go                     # FindRoot / PathPrefix / CurrentBranch
  render/                       # HTML / SVG / Markdown レンダリング
    svg.go                      # RenderSVG / RenderSVGWithBadges(…, updated) → SVG string
    html.go                     # RenderRoadmapPage / RenderIndexPage → HTML string
    drafts.go                   # Drafts: 下書きノードの扱い (公開ビルド=非活性 / プレビュー=バッジのみ)
    changelog.go                # 改版履歴 (ヘッダー / パネル) の HTML 断片
    markdown.go                 # RenderMarkdown / RenderMarkdownWithBase(body, urlPrefix) → HTML string (goldmark + chroma)
    links.go                    # RenderLinks([]Link) → HTML fragment
    theme.go                    # DeriveColors(hex) → {Base, Light}
    analytics.go                # RenderAnalyticsHead(Analytics) → <head> 用解析タグ
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
  progress_test.go              # 進捗率の下書き除外 (TEST:PROGRESS_BEGIN〜END) の検証
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

### 「この記事を編集」リンク (`editPath`)
- ビルド時に command 層 (`buildEditPaths`) が各ノードの記事のリポジトリルート相対パスを算出し、`NodeMeta.EditPath` → `ROADMAP_DATA[nodeId].editPath` で app.js に渡す
- 実ファイルは `content.Doc.RelPath` (content/ 起点・拡張子込み)、リポジトリ prefix は `repo.PathPrefix(configDir)` (`.git` ディレクトリ/ファイルを上位探索。無ければ空)
- 記事未作成ノードは `<prefix>content/<content or id>.md` (新規作成先の推定) を出す

### deploy の対象ブランチ
- `--branch/-b` 優先、未指定なら `repo.CurrentBranch` で `.git/HEAD` を読む (worktree/submodule の `gitdir:` も辿る)。detached HEAD・取得失敗時は `main` にフォールバックして表示
- ブランチ名は `[A-Za-z0-9._/-]+` のみ許可 (CI の YAML へそのまま埋め込むため)

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

### 下書きノード (`draft`)

- 指定は `roadmap.yml` のノードの `draft: true` と content frontmatter の `draft: true` の OR。`command.resolveDrafts(g, docs)` が ID 集合を返し、`buildNodeHTML` / `buildEditPaths` / `resolveNodeDocs` が参照する
- 公開ビルド (既定) は非活性表示: SVG は `<g class="roadmap-node is-draft" data-draft="1" aria-disabled="true">` + `<title>準備中</title>` + 「準備中」ラベル、接続エッジは `roadmap-edge is-draft` (CSS で破線・薄表示)。レイアウト座標は変えない
- **漏洩防止 (最重要)**: 公開ビルドの下書きは `NodeMeta` の html / text / links / editPath を空にして `draft: true` のみ渡す。本文の固有文字列が HTML に出ないことを `internal/render/drafts_test.go` で検証している。`render.Drafts{IDs, Preview}` の `locked` (公開) / `marked` (プレビュー) で分岐する
- `app.js` は `isLocked(id)` (`nodeData[id].draft && !SITE_CONFIG.draftPreview`) で `openPanel`・進捗の分母 (`isRequiredNode`)・検索・j/k ナビ・`updateNodeVisuals` から除外し、関連ノード一覧は「準備中」のリンクなし表示にする。localStorage / progressSync の既存進捗は削除しない (計算で無視するだけ)。index ページの分母は Go 側 (`graphNodeOrder` の exclude) で除外する
- 改版履歴: 下書きの frontmatter は `resolveNodeDocs` の入力から除外し、roadmap.yml の `changelog[].nodes` からの下書き参照は `changelog.StripNodes` で外す (項目自体は残す)。`validate` は `changelog.CheckDrafts` で「nodes が下書きだけ」の項目を警告し、記事未作成の下書きは未解決警告の対象外
- プレビュー: `runBuild(..., includeDrafts)` が true (`build --drafts` / `dev`) のとき `Preview: true` を渡し、`is-draft-preview` + 「下書き」バッジ (`node-draft-badge`) のみ付けて通常ノードとして出力する。`NodeMeta.Draft` は付けたまま `SITE_CONFIG.draftPreview: true` でフロントが区別する
- build 終了時に `printDrafts` が「下書き N 件 (非公開 / プレビュー表示)」と ID を表示する

### アクセス解析 (`site.analytics`)
- `provider` (umami / plausible / goatcounter / custom) が空なら無効。`render.RenderAnalyticsHead` が `<head>` 用タグを生成する (属性は `html/template` でエスケープ、custom の `head` のみ素通し)
- アダプタ方式: provider ごとに `window.roadmapperTrack(name, data)` を head で定義し、`app.js` は `track()` 経由でのみ呼ぶ。`SITE_CONFIG.analyticsEvents` が false なら何も送らず、例外は握りつぶす
- イベント: `node_open` / `node_state` / `share` / `outbound`。**deviceId・進捗データ・`?p=` の値は送らない**。シェアビューでは `node_state` を送らない
- `roadmapper dev` は常に解析を無効化 (`runBuild` の `noAnalytics`)、`build --no-analytics` でも無効化できる

## 禁止事項

- **外部バイナリ依存の追加禁止** — Node.js, graphviz, Python, etc. はインストール不要のまま保つ
- **フロントエンドフレームワーク追加禁止** — `web/static/app.js` は素の JS のまま維持する (目標 32KB 以内。現状 約 31.5KB)
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
- `internal/command/` の統合テストは手動確認。純粋なヘルパ (`buildEditPaths`, `resolveBranch`) のみユニットテストあり
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
