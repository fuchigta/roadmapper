---
links:
  - { title: "GitHub Actions ドキュメント", url: "https://docs.github.com/ja/actions" }
---

## 学ぶこと

CI (継続的インテグレーション) はコードの変更ごとに**自動でビルドとテスト**を行い、
CD (継続的デリバリー) はそれを**自動でデプロイ**まで広げる仕組みです。

## サブタスク

- [ ] CI と CD の違いを説明できる
- [ ] `.github/workflows/` にワークフローを書ける
- [ ] push や pull request をトリガーにテストを自動実行できる
- [ ] シークレットを安全に扱える

## ポイント

```yaml
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: go test ./...
```
