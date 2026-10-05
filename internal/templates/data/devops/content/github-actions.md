---
links:
  - { title: "GitHub Actions ドキュメント", url: "https://docs.github.com/ja/actions" }
---

## 学ぶこと

GitHub Actions は GitHub に組み込まれた CI/CD サービスです。
リポジトリ内の YAML ファイルでワークフローを定義します。

## サブタスク

- [ ] `.github/workflows` にワークフローを書ける
- [ ] `on` / `jobs` / `steps` の役割を説明できる
- [ ] シークレットを安全に扱える
- [ ] 既存アクションを `uses` で利用できる

## ポイント

```yaml
# .github/workflows/ci.yml
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7 # メジャーバージョンは公式リポジトリで最新を確認する
      - run: make test
```
