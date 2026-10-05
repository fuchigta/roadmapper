---
links:
  - { title: "Sass ドキュメント", url: "https://sass-lang.com/documentation/" }
---

## 学ぶこと

Sass (SCSS) は CSS を拡張した記法で、変数・ネスト・ミックスインなどを使えます。
最終的には通常の CSS にコンパイルされます。

## サブタスク

- [ ] 変数とネストを使ってスタイルを書ける
- [ ] パーシャルと `@use` でファイルを分割できる
- [ ] ミックスインを定義して再利用できる

## ポイント

```scss
$brand: #0ea5e9;
.btn {
  color: $brand;
  &:hover { opacity: 0.8; }
}
```

近年は CSS 自体にも変数やネストがあります。必要性を見極めて選びましょう。
