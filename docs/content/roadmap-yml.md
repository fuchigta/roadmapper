## roadmap.yml の全体構造

`roadmap.yml` は **サイト設定** と **ロードマップ定義** の 2 つのトップレベルキーで構成されます。

```yaml
site:
  title: My Roadmap
  description: ロードマップの説明文
  brandColor: "#6366f1"
  basePath: /my-repo/
  siteUrl: https://your-name.github.io/my-repo

roadmaps:
  - id: main
    title: メインロードマップ
    description: 説明
    nodes:
      - id: start
        title: はじめる
        type: required
```

## 学ぶこと

- **site セクション** — タイトル・テーマカラー・公開 URL など、サイト全体の設定
- **ノード定義** — `id` / `title` の命名ルールと最小構成
- **ノードタイプ** — `required` / `optional` / `alternative` の違いと表示
- **親子関係と依存関係** — `children` / `parents` による DAG の表現
- **複数ロードマップ** — `roadmaps:` 配列を使ったマルチマップ構成
