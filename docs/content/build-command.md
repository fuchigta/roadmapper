## build コマンドの使い方

```bash
roadmapper build -c my-roadmap/roadmap.yml -o my-roadmap/dist
```

| オプション | 説明 |
|---|---|
| `-c` / `--config` | `roadmap.yml` のパス (必須) |
| `-o` / `--output` | 出力ディレクトリ (デフォルト: `dist`) |

## 生成されるファイル

```
dist/
├── index.html             # ロードマップ一覧ページ
├── <roadmap-id>/
│   └── index.html         # 各ロードマップページ
├── style.css              # CSS (テーマカラー適用済み)
├── app.js                 # 進捗管理 / サイドパネル JS
├── sitemap.xml            # siteUrl 設定時のみ生成
└── feed.rss               # siteUrl 設定時のみ生成
```

## ビルドの特性

- **ビルドは冪等** — 同じ入力からは同じ出力が生成される
- **外部ランタイム不要** — Node.js・graphviz 等は不要
- **インクリメンタルビルドなし** — 毎回 `dist/` を全生成する

## サブタスク

- [ ] `roadmapper build -c roadmap.yml -o dist` が成功した
- [ ] `dist/index.html` をブラウザで開いて表示を確認した
- [ ] `dist/<roadmap-id>/index.html` で各ロードマップが表示された
