---
links:
  - { title: "GitHub Pages ドキュメント", url: "https://docs.github.com/ja/pages" }
  - { title: "GitHub Actions ドキュメント", url: "https://docs.github.com/ja/actions" }
---

## GitHub Pages 公開の流れ

roadmapper は `deploy` コマンドで GitHub Actions ワークフローを自動生成します。
`main` ブランチへの push をトリガーに、ビルドからデプロイまでを自動実行します。

```
roadmapper deploy → .github/workflows/pages.yml 生成
     ↓
git push → Actions 起動 → roadmapper build → dist/ → Pages 公開
```

## 学ぶこと

- **deploy --target github で Actions ファイル生成** — 生成される workflow の構造と再生成時の動作
- **リポジトリの Pages 設定** — Settings → Pages の Source 切替と独自ドメイン設定
- **siteUrl で sitemap / RSS を有効化** — 公開 URL 設定後に生成される追加ファイル
