---
links:
  - { title: "Umami", url: "https://umami.is/docs" }
  - { title: "Plausible", url: "https://plausible.io/docs" }
  - { title: "GoatCounter", url: "https://www.goatcounter.com/help" }
---

## アクセス解析を入れる

`site.analytics` を設定すると、ビルドした全ページの `<head>` に解析タグが入ります。
Cookie を使わない umami / Plausible / GoatCounter を想定しています。`provider` が空なら何も出力されません。

```yaml
site:
  analytics:
    provider: umami
    scriptUrl: https://analytics.example.com/script.js
    siteId: 11111111-2222-3333-4444-555555555555
    domains: [your-name.github.io]
```

| provider | `siteId` の意味 |
|---|---|
| `umami` | `data-website-id` |
| `plausible` | `data-domain` (サイトのドメイン) |
| `goatcounter` | `data-goatcounter` (例: `https://xxx.goatcounter.com/count`) |
| `custom` | 未使用 (`head` の HTML をそのまま挿入) |

> **`domains` の指定を推奨** — umami で計測を許可するホスト名を限定し、`localhost` やフォーク先からの混入を防ぎます。

## GA4 など他のサービス

`provider: custom` と `head` を使うと、任意の HTML を `<head>` に挿入できます。

```yaml
site:
  analytics:
    provider: custom
    head: |
      <script async src="https://www.googletagmanager.com/gtag/js?id=G-XXXXXXX"></script>
      <script>
        window.dataLayer = window.dataLayer || [];
        function gtag(){dataLayer.push(arguments);}
        gtag('js', new Date());
        gtag('config', 'G-XXXXXXX');
        window.roadmapperTrack = function (name, data) { gtag('event', name, data); };
      </script>
```

`window.roadmapperTrack` を定義すると、下記のイベントも転送されます (省略可)。

## 送信されるイベント

| イベント | データ | タイミング |
|---|---|---|
| `node_open` | `roadmap`, `node` | ノードの記事パネルを開いた |
| `node_state` | `roadmap`, `node`, `state` | 進捗状態を変更した (シェアビューでは送らない) |
| `share` | `roadmap` | 共有ボタンを押した |
| `outbound` | `roadmap`, `node`, `url` | パネル内の外部リンクをクリックした |

`events: false` を指定するとイベント送信を止め、ページビューの計測だけにできます。

## プライバシーと開発時の扱い

- 匿名デバイス ID・進捗データ・共有 URL の `?p=...` の値は送信しません。
- umami では `excludeSearch` (既定 true) により、共有 URL のクエリ文字列を記録しません。
- `roadmapper dev` では解析は常に無効です。本番ビルドでも外したいときは `roadmapper build --no-analytics` を使います。

## サブタスク

- [ ] `site.analytics` に provider / scriptUrl / siteId を設定した
- [ ] umami の場合は `domains` で公開ホスト名を指定した
- [ ] `build` 後の `dist/index.html` の `<head>` に解析タグが入っていることを確認した
- [ ] `dev` では解析タグが入らないことを確認した
