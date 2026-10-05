---
links:
  - { title: "Pro Git (日本語版)", url: "https://git-scm.com/book/ja/v2" }
  - { title: "Git ドキュメント", url: "https://git-scm.com/docs" }
---

## 学ぶこと

日々の開発で使う Git の基本操作です。まずは手を動かして一連の流れを身につけましょう。

## サブタスク

- [ ] `git add` / `git commit` で変更を記録できる
- [ ] ブランチを作成・切り替えできる
- [ ] `git merge` を実行しコンフリクトを解消できる
- [ ] `git log` / `git diff` で履歴と差分を確認できる

## ポイント

```bash
git switch -c feature/login
git add .
git commit -m "ログイン画面を追加"
git switch main
git merge feature/login
```
