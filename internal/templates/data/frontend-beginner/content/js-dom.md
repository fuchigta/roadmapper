---
links:
  - { title: "MDN: DOM の紹介", url: "https://developer.mozilla.org/ja/docs/Web/API/Document_Object_Model/Introduction" }
  - { title: "MDN: イベント入門", url: "https://developer.mozilla.org/ja/docs/Learn/JavaScript/Building_blocks/Events" }
---

## 学ぶこと

DOM はブラウザが HTML を扱うためのオブジェクトの集まりです。JavaScript から
要素を取得・変更し、クリックなどのイベントに反応させます。

## サブタスク

- [ ] `querySelector` で要素を取得できる
- [ ] テキストやクラスを書き換えられる
- [ ] `addEventListener` でイベントを処理できる
- [ ] 要素を動的に追加・削除できる

## ポイント

```js
document.querySelector("button").addEventListener("click", () => {
  document.body.classList.toggle("dark");
});
```
