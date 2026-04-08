---
links:
  - { title: "GitHub Pages ドキュメント", url: "https://docs.github.com/ja/pages/getting-started-with-github-pages/configuring-a-publishing-source-for-your-github-pages-site" }
---

## リポジトリの Pages 設定手順

1. GitHub リポジトリの **Settings** タブを開く
2. 左メニューの **Pages** をクリック
3. **Build and deployment** → **Source** を **GitHub Actions** に変更する
4. 保存後、`main` ブランチに push すると Actions が起動してデプロイが実行される

## 注意点

- Source が **Deploy from a branch** のままだと Actions 経由のデプロイが動きません
- デプロイ完了後、公開 URL は `https://<username>.github.io/<repository>/` になります

## 独自ドメインの設定

カスタムドメインを使う場合は Pages 設定の **Custom domain** に入力します。
その場合は `roadmap.yml` の `basePath` を `/` に変更してください。

```yaml
site:
  basePath: /          # カスタムドメインの場合
  siteUrl: https://example.com
```

## サブタスク

- [ ] Settings → Pages の Source を GitHub Actions に変更した
- [ ] `main` に push してデプロイが成功した
- [ ] 公開 URL でサイトが表示されることを確認した
