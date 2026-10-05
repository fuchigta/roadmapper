---
links:
  - { title: "Pro Git (日本語)", url: "https://git-scm.com/book/ja/v2" }
  - { title: "Git リファレンス", url: "https://git-scm.com/docs" }
---

## 学ぶこと

コミットは変更の**スナップショット**、ブランチはそこから分岐する作業の流れです。
この 2 つとマージの仕組みを理解すれば、日常の Git 操作の大半をこなせます。

## サブタスク

- [ ] `git add` / `git commit` で変更を記録できる
- [ ] `git log` / `git diff` で履歴と差分を確認できる
- [ ] `git switch -c` でブランチを作成・切り替えできる
- [ ] `git merge` でブランチを統合し、コンフリクトを解消できる

## ポイント

```bash
git switch -c feature/login
git add .
git commit -m "ログイン画面を追加"
git switch main
git merge feature/login
```
