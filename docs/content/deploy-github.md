---
links:
  - { title: "GitHub Actions ドキュメント", url: "https://docs.github.com/ja/actions" }
---

## deploy コマンドの実行

```bash
roadmapper deploy --target github              # 現在のブランチを対象にする
roadmapper deploy --target github --branch master  # ブランチを明示する
```

`.github/workflows/pages.yml` が生成されます。トリガーのブランチは `--branch` / `-b` で指定したもの、未指定ならカレントリポジトリの現在のブランチ (`.git/HEAD` から取得。detached HEAD や取得失敗時は `main` にフォールバックし、その旨を表示) になります。
出力先はカレントディレクトリ基準で固定のため、リポジトリのルートで実行してください。
`-c` / `-o` のようなオプションはありません。

## 生成されるワークフロー

```yaml
name: Deploy to GitHub Pages
on:
  push:
    branches: [main]   # 実際は --branch または現在のブランチ名
permissions:
  contents: read
  pages: write
  id-token: write
concurrency:
  group: pages
  cancel-in-progress: true
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install roadmapper
        run: |
          curl -sSfL https://github.com/fuchigta/roadmapper/releases/latest/download/roadmapper-linux-amd64.tar.gz \
            | tar -xz -C /usr/local/bin roadmapper
          chmod +x /usr/local/bin/roadmapper
      - name: Build
        run: roadmapper build --base "/${{ github.event.repository.name }}/"
      - uses: actions/upload-pages-artifact@v3
        with:
          path: dist
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deploy.outputs.page_url }}
    steps:
      - id: deploy
        uses: actions/deploy-pages@v4
```

ポイントは次のとおりです。

- GitHub Releases の最新バイナリを取得するため、CI に Go のセットアップは不要
- `--base` にリポジトリ名が自動で渡される (`site.basePath` より優先)
- 対象ブランチ (`--branch` または生成時の現在のブランチ) への push で起動する。後から変える場合は `branches:` を書き換えるか、`--branch` を付けて再生成する
- `roadmap.yml` がリポジトリ直下にない場合は `roadmapper build -c <パス>` に書き換える

## 再生成時の動作

ファイルが既に存在する場合は、内容が同じなら何もせず、異なる場合は差分を表示して上書きするか確認します。
ワークフローを手動編集している場合は注意してください。

## サブタスク

- [ ] `roadmapper deploy --target github` を実行した
- [ ] `.github/workflows/pages.yml` が生成された
- [ ] ファイルをリポジトリに push した
