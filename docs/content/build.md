## build コマンド

静的サイトを `dist/` ディレクトリに生成します。

```bash
roadmapper build -c my-roadmap/roadmap.yml -o my-roadmap/dist
```

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

## basePath の設定

GitHub Pages のリポジトリ名サブディレクトリで公開する場合は `basePath` を設定します。

```yaml
site:
  basePath: /my-repo/      # https://user.github.io/my-repo/
```

ルートで公開する場合 (カスタムドメインなど) は空欄のままにします。

## サブタスク

- [ ] `build` コマンドで `dist/` が生成された
- [ ] `dist/index.html` をブラウザで開いて確認した
- [ ] `basePath` を自分のリポジトリ名に合わせた
