## ビルドの概要

`roadmapper build` は `roadmap.yml` とコンテンツ Markdown を読み込み、
外部ランタイム不要で完全な静的サイトを生成します。

生成されるファイル:

```
dist/
├── index.html          # ロードマップ一覧ページ
├── getting-started/    # 各ロードマップページ (id が名前)
│   └── index.html
├── style.css
├── app.js
├── sitemap.xml         # siteUrl が設定されている場合
└── feed.rss            # siteUrl が設定されている場合
```

## 学ぶこと

- **build コマンドで静的ファイル生成** — `-c` / `-o` オプションと出力ファイル一覧
- **basePath の設定** — サブディレクトリ公開とルート公開の使い分け
