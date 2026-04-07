# CLAUDE.md — roadmapper 開発ガイド

## プロジェクト概要

Go 製の学習ロードマップ静的サイトジェネレータ CLI。`roadmap.yml` + `content/*.md` から
GitHub Pages / GitLab Pages 対応の静的サイトを生成する。**外部ランタイム依存ゼロ** が最重要設計原則。

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
go run ./cmd/roadmapper validate -c demo/roadmap.yml
go run ./cmd/roadmapper dev -c demo/roadmap.yml   # Ctrl+C で終了
```

## ディレクトリ構造

```
cmd/roadmapper/main.go          # CLIエントリ (cobra)
internal/
  command/                      # CLI サブコマンド実装
    build.go                    # roadmapper build
    dev.go                      # roadmapper dev (fsnotify + SSE livereload)
    deploy.go                   # roadmapper deploy --target github|gitlab
    init.go                     # roadmapper init
    validate.go                 # roadmapper validate
  config/                       # roadmap.yml パーサ + バリデーション
    schema.go                   # Config / Site / Roadmap / Node / Link 構造体
    loader.go                   # Load(path) → *Config
    validate.go                 # Validate(*Config) error
  content/                      # content/<id>.md ローダ
    loader.go                   # LoadDir(dir) → map[string]*Doc
  graph/                        # ノード/エッジ DAG モデル
    graph.go                    # Build(*Roadmap) → *Graph (cycle detection)
  layout/                       # Goja + dagre.js でレイアウト計算
    layout.go                   # Compute(*Graph, *Config) → *Result
    vendor/dagre.min.js         # //go:embed で同梱
  meta/                         # OGP / sitemap / RSS
    sitemap.go                  # RenderSitemap(*Config) → XML string
    rss.go                      # RenderRSS(*Config, graphs) → XML string
  render/                       # HTML / SVG / Markdown レンダリング
    svg.go                      # RenderSVG(*Graph, *Result, brandColor) → SVG string
    html.go                     # RenderRoadmapPage / RenderIndexPage → HTML string
    markdown.go                 # RenderMarkdown(body) → HTML string (goldmark + chroma)
    links.go                    # RenderLinks([]Link) → HTML fragment
    theme.go                    # DeriveColors(hex) → {Base, Light}
  server/                       # dev サーバ
    server.go                   # HTTP server + SSE + livereload script injection
  templates/                    # init コマンド用スケルトン
    embed.go                    # //go:embed all:data
    data/minimal/               # 最小サンプル
    data/frontend-beginner/     # 現実的テンプレート
web/                            # ビルド時埋め込みアセット
  embed.go                      # //go:embed templates static
  templates/index.html          # インデックスページ
  templates/roadmap.html        # ロードマップページ
  static/style.css              # CSS variables ベースのテーマ
  static/app.js                 # 進捗トラッキング / サイドパネル / テーマ切替
```

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

## 禁止事項

- **外部バイナリ依存の追加禁止** — Node.js, graphviz, Python, etc. はインストール不要のまま保つ
- **フロントエンドフレームワーク追加禁止** — `web/static/app.js` は素の JS のまま維持する (目標 30KB 以内)
- **`filepath.Join` を embed FS パスに使用禁止** — `path.Join` を使うこと
- **`web/` 以下のファイルをビルド外から直接コピーしない** — `web.FS` 経由でアクセスする

## テスト方針

- 各 `internal/` パッケージにユニットテストを置く
- ゴールデンファイルテストは `testdata/` ディレクトリに配置
- `go test ./...` がすべて通ること
- `internal/command/` にはテストファイルなし (統合テストは手動確認)

## 依存ライブラリ

| ライブラリ | 用途 |
|---|---|
| `github.com/spf13/cobra` | CLI フレームワーク |
| `github.com/dop251/goja` | Pure Go JS エンジン (dagre.js 実行) |
| `github.com/yuin/goldmark` | Markdown → HTML |
| `github.com/alecthomas/chroma/v2` | シンタックスハイライト |
| `github.com/fsnotify/fsnotify` | ファイル監視 (dev コマンド) |
| `gopkg.in/yaml.v3` | YAML パース |
