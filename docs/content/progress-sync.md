---
links:
  - { title: "Cloudflare Workers KV", url: "https://developers.cloudflare.com/kv/" }
---

## 進捗のバックエンド同期

進捗は通常ブラウザの localStorage にだけ保存されます。
`progressSync` を設定すると、複数デバイス間で進捗を同期できます。
同期先のサーバは **サイト作者が用意します** (roadmapper はクライアント側のコードだけを生成します)。

```yaml
site:
  progressSync:
    enabled: true
    endpoint: https://api.example.com/sync   # 末尾スラッシュなし
```

`enabled: true` のときは `endpoint` が必須で、`http://` または `https://` で始まる必要があります
(満たさないと `validate` がエラーにします)。

## HTTP コントラクト

| メソッド | パス | 概要 |
|---|---|---|
| `GET` | `{endpoint}/{deviceId}/{roadmapId}` | ロードマップ進捗を取得 |
| `PUT` | `{endpoint}/{deviceId}/{roadmapId}` | ロードマップ進捗を保存 |

- `deviceId` はブラウザ初回アクセス時に自動生成される匿名 UUID
- `GET` が `404` を返すと「データなし」として扱われ、その他の失敗は無視されます
- `PUT` のボディ: `{ "<nodeId>": { "state": "done", "tasks": [true, false] }, ... }`
- サーバは冪等な全置換で実装すれば十分です

## 同期の動き

- ページ表示時にリモート進捗を取得し、ローカルとマージします。ノードごとに進んでいる方を採用します (`none` < `in-progress` < `done` / `skipped`、同ランクなら `done` 優先)。チェックリストは OR で結合します
- 状態を変更すると、800ms 後にまとめて `PUT` します
- オフライン等で失敗した場合は、`online` イベントか次回表示時に再送します
- 共有 URL (`?p=...`) で開いた読み取り専用ビューでは同期しません

## CORS 設定 (必須)

`PUT` は `Content-Type: application/json` を付けるため、サーバは preflight (`OPTIONS`) に応答する必要があります。

```
Access-Control-Allow-Origin: https://your-site.example.com
Access-Control-Allow-Methods: GET, PUT, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

認証はリバースプロキシや Cloudflare Access などで付与してください。

## サブタスク

- [ ] 同期用のサーバ (GET / PUT / OPTIONS) を用意した
- [ ] `progressSync.enabled` と `endpoint` を設定した
- [ ] 別ブラウザで開いて進捗が同期されることを確認した
- [ ] CORS ヘッダーが正しく返ることを確認した
