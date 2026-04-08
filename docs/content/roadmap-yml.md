## roadmap.yml の全体構造

```yaml
site:
  title: My Roadmap
  description: ロードマップの説明文
  brandColor: "#6366f1"   # メインカラー (HEX)
  author: your-name
  basePath: /my-repo/     # GitHub Pages のサブディレクトリ
  siteUrl: https://your-name.github.io/my-repo

roadmaps:
  - id: main
    title: メインロードマップ
    description: 説明
    nodes:
      - id: start
        title: はじめる
        type: required
        children:
          - id: step1
            title: ステップ 1
```

## site セクション

| キー | 説明 |
|---|---|
| `title` | サイトタイトル |
| `description` | メタ description |
| `brandColor` | テーマカラー (HEX) |
| `basePath` | GitHub Pages サブディレクトリ (例: `/my-repo/`) |
| `siteUrl` | 公開 URL。設定すると sitemap.xml と RSS が生成される |
| `author` | 著者名 |
| `repo` | リポジトリ URL (「ソースを見る」リンクに使用) |
| `editBranch` | 編集リンクのブランチ名 |

## ノード定義

| キー | 説明 |
|---|---|
| `id` | ノードの一意 ID |
| `title` | ノード名 |
| `type` | `required` / `optional` / `alternative` |
| `children` | 子ノードのリスト |
| `parents` | 追加の親ノード ID リスト (複数親の DAG) |

## サブタスク

- [ ] site セクションを自分の情報に書き換えた
- [ ] ノードの id と title を定義した
- [ ] type を設定した
- [ ] validate でエラーゼロを確認した
