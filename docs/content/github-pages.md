---
links:
  - { title: "GitHub Pages ドキュメント", url: "https://docs.github.com/ja/pages" }
  - { title: "GitHub Actions ドキュメント", url: "https://docs.github.com/ja/actions" }
---

## deploy コマンドで GitHub Actions ファイルを生成

```bash
roadmapper deploy --target github -c my-roadmap/roadmap.yml
```

`.github/workflows/pages.yml` が生成されます。
既存のファイルがある場合は差分を表示して上書きするか確認します。

生成される Actions ワークフロー:

```yaml
name: Deploy to GitHub Pages
on:
  push:
    branches: [main]
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: go install github.com/fuchigta/roadmapper/cmd/roadmapper@latest
      - run: roadmapper build -c roadmap.yml -o dist
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist
      - uses: actions/deploy-pages@v4
```

## GitHub リポジトリの Pages 設定

1. GitHub リポジトリの **Settings → Pages** を開く
2. Source を **GitHub Actions** に設定する
3. `main` ブランチに push すると自動デプロイが実行される

## siteUrl の設定

公開 URL が確定したら `roadmap.yml` の `siteUrl` を設定します。

```yaml
site:
  siteUrl: https://your-name.github.io/my-repo
```

これで `sitemap.xml` と `feed.rss` が生成されるようになります。

## サブタスク

- [ ] `roadmapper deploy --target github` で Actions ファイルを生成した
- [ ] GitHub リポジトリの Settings → Pages で Source を GitHub Actions に設定した
- [ ] `main` に push してデプロイが成功した
- [ ] `siteUrl` を設定して sitemap.xml の生成を確認した
