---
links:
  - { title: "Cloudflare: DNS とは", url: "https://www.cloudflare.com/learning/dns/what-is-dns/" }
  - { title: "RFC 1035", url: "https://www.rfc-editor.org/rfc/rfc1035" }
---

## 学ぶこと

DNS はドメイン名を IP アドレスに変換する仕組みです。
名前解決の失敗は障害原因として非常に多いため、動きを理解しておくことが重要です。

## サブタスク

- [ ] A / AAAA / CNAME / MX / TXT レコードの役割を説明できる
- [ ] 名前解決の流れ (リゾルバ→権威サーバー) を説明できる
- [ ] `dig` で名前解決を確認できる
- [ ] TTL がキャッシュに与える影響を理解している

## ポイント

```bash
dig example.com A
dig +short example.com NS
```
