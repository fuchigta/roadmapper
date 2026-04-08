---
links:
  - { title: "GitHub Actions ドキュメント", url: "https://docs.github.com/ja/actions" }
---

## deploy コマンドの実行

```bash
roadmapper deploy --target github -c my-roadmap/roadmap.yml
```

`.github/workflows/pages.yml` が生成されます。

## 生成されるワークフロー

```yaml
name: Deploy to GitHub Pages
on:
  push:
    branches: [main]
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pages: write
      id-token: write
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

## 再生成時の動作

ファイルが既に存在する場合は差分を表示して上書きするか確認します。
ワークフローを手動編集している場合は注意してください。

## サブタスク

- [ ] `roadmapper deploy --target github` を実行した
- [ ] `.github/workflows/pages.yml` が生成された
- [ ] ファイルをリポジトリに push した
