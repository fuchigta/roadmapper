---
links:
  - { title: "Bash リファレンスマニュアル (GNU)", url: "https://www.gnu.org/software/bash/manual/" }
  - { title: "ShellCheck", url: "https://www.shellcheck.net/" }
---

## 学ぶこと

繰り返し行う作業をシェルスクリプトにまとめると、手作業のミスを減らし再現性を高められます。
自動化の第一歩として、短いスクリプトを書く習慣をつけましょう。

## サブタスク

- [ ] 変数・条件分岐・ループを使ったスクリプトを書ける
- [ ] 引数 (`$1`, `$@`) と終了ステータス (`$?`) を扱える
- [ ] `set -euo pipefail` の意味を説明できる
- [ ] ShellCheck で警告を確認する習慣がある

## ポイント

```bash
#!/usr/bin/env bash
set -euo pipefail

for f in *.log; do
  gzip "$f"
done
```
