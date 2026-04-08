## site セクション

`site:` はサイト全体に関わるメタ情報を定義します。

```yaml
site:
  title: My Roadmap          # ブラウザタイトル・OGP タイトル
  description: 説明文         # メタ description
  brandColor: "#6366f1"      # テーマカラー (HEX 6 桁)
  author: your-name          # 著者名 (フッターに表示)
  repo: https://github.com/you/repo   # 「ソースを見る」リンク
  editBranch: main           # コンテンツ編集リンクのブランチ名
  basePath: /my-repo/        # サブディレクトリ公開時のパス
  siteUrl: https://you.github.io/my-repo  # sitemap/RSS に使用
```

## キー詳細

| キー | 必須 | 説明 |
|---|---|---|
| `title` | 推奨 | サイトのタイトル |
| `description` | 任意 | meta description |
| `brandColor` | 任意 | HEX 6 桁 (例: `#3b82f6`)。未設定時はデフォルトカラー |
| `author` | 任意 | フッターに表示される著者名 |
| `repo` | 任意 | 設定するとヘッダーにリポジトリリンクが表示される |
| `editBranch` | 任意 | `repo` と組み合わせて各ノードの「このページを編集」リンクを生成 |
| `basePath` | 任意 | GitHub Pages のリポジトリサブディレクトリ (例: `/my-repo/`) |
| `siteUrl` | 任意 | 設定すると `sitemap.xml` と `feed.rss` が生成される |

## サブタスク

- [ ] `title` と `description` を設定した
- [ ] `brandColor` を自分好みのカラーに変更した
- [ ] `basePath` を公開先に合わせて設定した
- [ ] `validate` コマンドでエラーゼロを確認した
