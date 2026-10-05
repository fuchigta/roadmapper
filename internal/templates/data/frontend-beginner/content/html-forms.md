---
links:
  - { title: "MDN: Web フォーム", url: "https://developer.mozilla.org/ja/docs/Learn/Forms" }
  - { title: "MDN: フォームデータのバリデーション", url: "https://developer.mozilla.org/ja/docs/Learn/Forms/Form_validation" }
---

## 学ぶこと

フォームはユーザーからの入力を受け取る基本の仕組みです。HTML の属性だけでも
基本的な入力チェックができます。

## サブタスク

- [ ] `<label>` と `<input>` を `for` / `id` で関連付けられる
- [ ] `type` (text / email / number / checkbox / radio など) を使い分けられる
- [ ] `required`・`pattern`・`min` / `max` で入力チェックができる
- [ ] `<select>`・`<textarea>`・`<button>` を使える

## ポイント

```html
<label for="mail">メールアドレス</label>
<input id="mail" type="email" name="mail" required>
```

ブラウザ側のチェックは回避できるため、サーバー側の検証も必ず必要です。
