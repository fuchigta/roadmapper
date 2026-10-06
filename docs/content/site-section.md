## site セクション

`site:` はサイト全体に関わるメタ情報を定義します。

```yaml
site:
  title: My Roadmap          # ブラウザタイトル・OGP タイトル (必須)
  description: 説明文         # メタ description
  brandColor: "#6366f1"      # テーマカラー (HEX 6 桁)
  author: your-name          # 著者名 (トップページのフッターに表示)
  license: CC-BY-4.0         # ライセンス表記 (author の隣に表示)
  repo: https://github.com/you/repo   # 「この記事を編集」リンクの生成元
  editBranch: main           # 編集リンクのブランチ名
  basePath: /my-repo/        # サブディレクトリ公開時のパス
  siteUrl: https://you.github.io/my-repo  # sitemap / RSS / og:url に使用
```

## キー詳細

| キー | 必須 | 既定値 | 説明 |
|---|---|---|---|
| `title` | **必須** | — | サイトのタイトル。空だと `validate` / `build` がエラーになる |
| `description` | 任意 | — | meta description |
| `brandColor` | 任意 | `#4f46e5` | HEX 6 桁 (例: `#3b82f6`)。アクセントカラーとその淡色が CSS 変数として自動派生される |
| `author` | 任意 | — | トップページのフッターに `by <author>` と表示される |
| `license` | 任意 | — | `author` を設定したときに限りフッターに表示される |
| `repo` | 任意 | — | 設定するとサイドパネルに「この記事を編集」リンクが表示される (GitLab の URL なら GitLab 形式の編集 URL) |
| `editBranch` | 任意 | `main` | `repo` と組み合わせて編集リンクを生成する |
| `basePath` | 任意 | 空 | GitHub Pages のリポジトリサブディレクトリ (例: `/my-repo/`) |
| `siteUrl` | 任意 | 空 | 設定すると `sitemap.xml` / `feed.rss` が生成され、`og:url` が付く |

このほかのキーは別の記事で扱います。

| キー | 参照 |
|---|---|
| `layout` (`rankDir` / `nodeSep` / `rankSep`)、`panel` (`width` / `minWidth` / `maxWidth`) | レイアウトとパネル幅 |
| `contentAssets` (`exclude`) | 画像など静的ファイルの参照 |
| `progressSync` (`enabled` / `endpoint`) | 進捗のバックエンド同期 |
| `analytics` (`provider` / `scriptUrl` / `siteId` など) | アクセス解析 |

## サブタスク

- [ ] `title` と `description` を設定した
- [ ] `brandColor` を自分好みのカラーに変更した
- [ ] `basePath` を公開先に合わせて設定した
- [ ] `validate` コマンドでエラーゼロを確認した
