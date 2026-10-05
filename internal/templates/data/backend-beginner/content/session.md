---
links:
  - { title: "MDN: HTTP クッキーの使用", url: "https://developer.mozilla.org/ja/docs/Web/HTTP/Cookies" }
---

## 学ぶこと

HTTP はステートレスなので、ログイン状態を保つには工夫が必要です。
サーバーがセッション情報を保持し、**Cookie** のセッション ID で利用者を識別するのが古典的な方法です。

## サブタスク

- [ ] Cookie の仕組みと `Set-Cookie` ヘッダーを説明できる
- [ ] `HttpOnly` / `Secure` / `SameSite` 属性の意味を説明できる
- [ ] セッション ID の保存先 (メモリ・DB・Redis など) を比較できる
- [ ] CSRF とは何かを説明できる
