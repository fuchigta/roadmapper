---
links:
  - { title: "MDN: HTML の基礎", url: "https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/HTML_basics" }
  - { title: "HTML Living Standard", url: "https://html.spec.whatwg.org/" }
---

## 学ぶこと

HTML はウェブページの**骨格**を作る言語です。どの要素がどんな意味を持つのかを
理解するところから始めましょう。

## サブタスク

- [ ] `<header>` / `<main>` / `<footer>` / `<article>` / `<section>` を使い分けられる
- [ ] リンク・画像・テーブル・リストを正しく書ける
- [ ] `<form>` と各種 `<input>` type を使ってフォームを作れる
- [ ] HTML バリデーター (validator.w3.org) でエラーゼロにできる
- [ ] alt 属性やランドマークロールなど基本的なアクセシビリティを理解している

## ポイント

```html
<!-- セマンティックな構造の例 -->
<article>
  <h2>記事タイトル</h2>
  <p>本文...</p>
</article>
```

> `<div>` と `<span>` はあくまで意味のない汎用コンテナ。
> 適切なセマンティック要素を優先しましょう。
