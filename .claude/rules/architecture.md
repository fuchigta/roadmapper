# アーキテクチャ原則

## シングルバイナリ原則

**外部ランタイム依存をゼロに保つ** — これが最重要の差別化ポイント。

- Node.js, graphviz, Python, Ruby, pandoc などを実行時に必要としてはならない
- 新しいライブラリを追加するときは Pure Go 実装を優先する
- バイナリサイズの目安: 20MB 以内 (Goja + dagre.js 込み)

## データフロー

```
roadmap.yml ──▶ config.Load() ──▶ graph.Build() ──▶ layout.Compute()
                                                            │
content/*.md ─▶ content.LoadDir() ─▶ render.RenderMarkdown()
                                                            │
                                              render.RenderSVG()
                                              render.RenderRoadmapPage()
                                              meta.RenderSitemap()
                                              meta.RenderRSS()
                                                            │
                                                         dist/
```

各ステップは純粋関数として実装する。グローバル状態は持たない。

## レイヤー責務

| レイヤー | 責務 | 禁止事項 |
|---|---|---|
| `internal/config` | YAML 読み込みと検証のみ | ファイル出力、ネットワーク |
| `internal/graph` | DAG 構築と検証のみ | 座標計算、レンダリング |
| `internal/layout` | 座標計算のみ | ファイル出力、HTML 生成 |
| `internal/render` | 文字列生成のみ | ファイル出力、ネットワーク |
| `internal/meta` | XML 文字列生成のみ | ファイル出力 |
| `internal/command` | 上記を組み合わせてファイル出力 | ビジネスロジック |
| `internal/server` | HTTP 配信のみ | ビルドロジック |

## フロントエンド方針

- `web/static/app.js` は素の JavaScript のみ (React/Vue 等のフレームワーク禁止)
- 目標ファイルサイズ: 30KB 以内
- `web/static/style.css` は CSS カスタムプロパティ (変数) ベース
- ビルド時に HTML へ静的埋め込みできるものは埋め込む (JS は最小限に)

## 拡張ポイント

将来の拡張を想定した設計箇所:
- `config.Site.Layout` — dagre パラメータを公開している (rankDir, nodeSep, rankSep)
- `config.Node.X`, `config.Node.Y` — 手動座標オーバーライド
- `config.Site.SiteURL` — OGP/sitemap/RSS は siteUrl が空なら生成しない (オプション機能)
- `content.Doc.Frontmatter.Links` — frontmatter のリンクが roadmap.yml より優先される
