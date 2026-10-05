---
links:
  - { title: "MDN: HTTP の概要", url: "https://developer.mozilla.org/ja/docs/Web/HTTP/Overview" }
  - { title: "MDN: HTTP レスポンスステータスコード", url: "https://developer.mozilla.org/ja/docs/Web/HTTP/Status" }
---

## 学ぶこと

HTTP はウェブ上でクライアントとサーバーがやり取りするための**プロトコル**です。
HTTPS は HTTP を TLS で暗号化したもので、現在ではほぼ必須です。

## サブタスク

- [ ] リクエストとレスポンスの構造 (メソッド・ヘッダー・ボディ) を説明できる
- [ ] GET / POST / PUT / DELETE の使い分けを説明できる
- [ ] 主なステータスコード (200, 301, 400, 401, 404, 500) の意味を言える
- [ ] HTTPS が何を守ってくれるのか説明できる

## ポイント

```http
GET /users/1 HTTP/1.1
Host: example.com
Accept: application/json
```
