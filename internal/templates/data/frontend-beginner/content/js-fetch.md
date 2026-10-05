---
links:
  - { title: "MDN: Fetch API", url: "https://developer.mozilla.org/ja/docs/Web/API/Fetch_API" }
  - { title: "MDN: Fetch の使用", url: "https://developer.mozilla.org/ja/docs/Web/API/Fetch_API/Using_Fetch" }
---

## 学ぶこと

`fetch` を使って、サーバーの API から JSON などのデータを取得・送信します。

## サブタスク

- [ ] GET リクエストで JSON を取得して画面に表示できる
- [ ] POST でデータを送信できる
- [ ] `res.ok` でエラー応答を判定できる
- [ ] CORS エラーが出る理由を説明できる

## ポイント

`fetch` は 404 や 500 でも reject されません。`res.ok` の確認を忘れないようにしましょう。
