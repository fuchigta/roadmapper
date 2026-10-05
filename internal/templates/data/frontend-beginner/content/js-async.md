---
links:
  - { title: "MDN: 非同期 JavaScript", url: "https://developer.mozilla.org/ja/docs/Learn/JavaScript/Asynchronous" }
  - { title: "MDN: Promise", url: "https://developer.mozilla.org/ja/docs/Web/JavaScript/Reference/Global_Objects/Promise" }
---

## 学ぶこと

通信のように時間のかかる処理を、画面を止めずに扱う方法を学びます。

## サブタスク

- [ ] 同期処理と非同期処理の違いを説明できる
- [ ] Promise の `then` / `catch` を使える
- [ ] `async` / `await` で書き換えられる
- [ ] `try` / `catch` でエラーを処理できる

## ポイント

```js
async function load() {
  try {
    const res = await fetch("/data.json");
    return await res.json();
  } catch (e) {
    console.error(e);
  }
}
```
