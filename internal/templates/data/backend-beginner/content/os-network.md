---
links:
  - { title: "curl 公式ドキュメント", url: "https://curl.se/docs/" }
  - { title: "Linux man-pages: ss(8)", url: "https://man7.org/linux/man-pages/man8/ss.8.html" }
---

## 学ぶこと

ネットワークの問題は、コマンドで切り分けられると解決が早くなります。
特に `curl` は API の動作確認に日常的に使います。

## サブタスク

- [ ] `curl` で GET / POST リクエストを送り、レスポンスヘッダーを確認できる
- [ ] `ping` で疎通確認ができる
- [ ] `netstat` または `ss` で待ち受け中のポートを確認できる
- [ ] `dig` で DNS の名前解決を確認できる

## ポイント

```bash
# JSON を POST し、レスポンスヘッダーも表示する
curl -i -X POST https://example.com/api/items \
  -H "Content-Type: application/json" \
  -d '{"name":"sample"}'
```
