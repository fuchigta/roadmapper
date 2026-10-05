---
links:
  - { title: "curl 公式ドキュメント", url: "https://curl.se/docs/" }
  - { title: "Go: net/http/httptest", url: "https://pkg.go.dev/net/http/httptest" }
---

## 学ぶこと

API テストは、HTTP リクエストを送って**ステータスコードやレスポンスの内容**を確認するテストです。
API の仕様どおりに動いているかを外側から検証します。

## サブタスク

- [ ] `curl` などでエンドポイントを手動確認できる
- [ ] テストコードから HTTP リクエストを送って検証できる
- [ ] 正常系だけでなく、不正な入力や認証エラーも確認できる
- [ ] レスポンスの JSON 構造を検証できる
