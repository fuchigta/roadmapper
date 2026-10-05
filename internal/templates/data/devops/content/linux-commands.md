---
links:
  - { title: "man7.org: Linux man pages", url: "https://man7.org/linux/man-pages/" }
  - { title: "GNU Coreutils マニュアル", url: "https://www.gnu.org/software/coreutils/manual/" }
---

## 学ぶこと

日常的な操作とトラブルシュートの大半は、基本コマンドの組み合わせで行えます。
パイプ (`|`) でつなぐ使い方に慣れることが、運用作業の効率を大きく左右します。

## サブタスク

- [ ] `ls` / `cd` / `cp` / `mv` / `rm` でファイルを操作できる
- [ ] `grep` でログから必要な行を検索できる
- [ ] `awk` / `sed` で簡単なテキスト整形ができる
- [ ] パイプとリダイレクト (`>`, `>>`, `2>&1`) を使い分けられる

## ポイント

```bash
# エラー行の件数を種類別に数える
grep ERROR app.log | awk '{print $4}' | sort | uniq -c | sort -nr
```
