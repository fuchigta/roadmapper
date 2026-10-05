---
links:
  - { title: "MDN: レスポンシブデザイン", url: "https://developer.mozilla.org/ja/docs/Learn/CSS/CSS_layout/Responsive_Design" }
  - { title: "MDN: メディアクエリーの使用", url: "https://developer.mozilla.org/ja/docs/Web/CSS/CSS_media_queries/Using_media_queries" }
---

## 学ぶこと

画面幅の異なるデバイスで読みやすく表示するための考え方と技術です。

## サブタスク

- [ ] `<meta name="viewport">` を設定できる
- [ ] メディアクエリでスタイルを切り替えられる
- [ ] `%`・`rem`・`vw` など相対的な単位を使い分けられる
- [ ] 画像が親からはみ出さないようにできる

## ポイント

```css
@media (min-width: 768px) {
  .layout { display: grid; grid-template-columns: 1fr 2fr; }
}
```

小さい画面を基準に書き、広い画面向けに追加する「モバイルファースト」が扱いやすいです。
