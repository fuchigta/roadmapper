## siteUrl の役割

`siteUrl` を設定すると、ビルド時に以下のファイルが追加生成されます。

| ファイル | 用途 |
|---|---|
| `sitemap.xml` | 検索エンジンへのインデックス通知 (トップと各ロードマップページの URL) |
| `feed.rss` | RSS リーダーへのコンテンツ配信 |
| `og:url` メタタグ | 各ページの OGP URL |

## 設定方法

```yaml
site:
  siteUrl: https://your-name.github.io/my-repo
```

> **末尾スラッシュは不要** — `https://...` 形式で記述します。

## 未設定時の挙動

`siteUrl` を設定しなくても静的サイト本体のビルドは正常に動作します。
sitemap と RSS は生成されず、`og:url` も付きません (`og:title` などの OGP タグは常に出力されます)。
不要な場合は空欄のままで問題ありません。

## 設定後の確認

```bash
roadmapper build -c roadmap.yml -o dist
ls dist/sitemap.xml dist/feed.rss   # 存在することを確認
```

URL は `siteUrl` + `basePath` (末尾スラッシュは自動補完) + `<roadmapId>/index.html` の形で組み立てられます。
RSS フィードの item は、改版履歴 (changelog) があるロードマップでは履歴項目ごと、なければノードごとに出力されます。
結合結果が正しいかを確認してください。

## サブタスク

- [ ] `siteUrl` に公開 URL を設定した
- [ ] `build` 後に `dist/sitemap.xml` が生成された
- [ ] `dist/feed.rss` が生成された
- [ ] RSS の URL が正しいことを確認した
