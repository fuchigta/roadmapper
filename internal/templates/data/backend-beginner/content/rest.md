---
links:
  - { title: "MDN: HTTP リクエストメソッド", url: "https://developer.mozilla.org/ja/docs/Web/HTTP/Methods" }
  - { title: "MDN: HTTP レスポンスステータスコード", url: "https://developer.mozilla.org/ja/docs/Web/HTTP/Status" }
---

## 学ぶこと

REST は HTTP を使って**リソース**を操作する API 設計のスタイルです。
URL でリソースを表し、HTTP メソッドで操作を表すのが基本です。

## サブタスク

- [ ] リソースを表す URL (名詞・複数形) を設計できる
- [ ] メソッドごとの意味 (GET / POST / PUT / PATCH / DELETE) を使い分けられる
- [ ] 適切なステータスコードを返せる
- [ ] ページネーションやエラーレスポンスの形式を決められる

## ポイント

```text
GET    /users        ユーザー一覧
POST   /users        ユーザー作成
GET    /users/1      ユーザー取得
DELETE /users/1      ユーザー削除
```
