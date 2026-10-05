---
links:
  - { title: "npm: package.json", url: "https://docs.npmjs.com/cli/configuring-npm/package-json" }
---

## 学ぶこと

`package.json` はプロジェクトの名前・スクリプト・依存関係を記録する設定ファイルです。
他人のプロジェクトを読むときの入口になります。

## サブタスク

- [ ] `dependencies` と `devDependencies` の違いを説明できる
- [ ] `scripts` に書かれたコマンドを読める
- [ ] バージョン指定 (`^` や `~`) の意味を説明できる
- [ ] ロックファイルの役割を理解している

## ポイント

```json
{
  "scripts": { "dev": "vite", "build": "vite build" },
  "devDependencies": { "vite": "^5.0.0" }
}
```
