## build コマンドの使い方

```bash
roadmapper build -c my-roadmap/roadmap.yml -o my-roadmap/dist
```

| オプション | 既定値 | 説明 |
|---|---|---|
| `-c` / `--config` | `roadmap.yml` | `roadmap.yml` のパス。`content/` はこのファイルと同じディレクトリを基準に読み込む |
| `-o` / `--out` | `dist` | 出力ディレクトリ |
| `--base` | 空 | ベースパス (例: `/my-repo/`)。指定すると `site.basePath` より優先される |
| `--no-analytics` | 無効 | アクセス解析タグを出力しない |
| `--drafts` | 無効 | 下書きノードを通常ノードとして出力し「下書き」バッジを付ける (ステージング用)。既定では下書きは非活性の「準備中」になり本文を出力しない (「下書きノード」の記事を参照) |

## 生成されるファイル

```
dist/
├── index.html             # ロードマップ一覧ページ
├── <roadmap-id>/
│   └── index.html         # 各ロードマップページ
├── content/               # content/ 内の画像など非 .md ファイル (あれば)
├── style.css              # CSS (テーマカラー適用済み)
├── app.js                 # 進捗管理 / サイドパネル JS
├── sitemap.xml            # siteUrl 設定時のみ生成
└── feed.rss               # siteUrl 設定時のみ生成
```

## ビルドの特性

- **ビルドは冪等** — 同じ入力からは同じ出力が生成される (ただし更新バッジは現在日時から判定する)
- **外部ランタイム不要** — Node.js・graphviz 等は不要
- **インクリメンタルビルドなし** — 毎回すべてのファイルを再生成する (`dist/` 内の古いファイルは削除されない)
- **content の警告** — ノードに対応する `content/*.md` が見つからないと warning を表示する (ビルドは続行)

## サブタスク

- [ ] `roadmapper build -c roadmap.yml -o dist` が成功した
- [ ] `dist/index.html` をブラウザで開いて表示を確認した
- [ ] `dist/<roadmap-id>/index.html` で各ロードマップが表示された
