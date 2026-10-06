---
links:
  - { title: "roadmapper README (content/ のサブディレクトリ)", url: "https://github.com/fuchigta/roadmapper#content-のサブディレクトリ" }
---

## サブディレクトリでコンテンツを整理する

`content/` 配下は再帰的にスキャンされるため、記事が増えたらフォルダに分けて管理できます。

```
content/
├── html.md             # ルート直下: ID "html" に対応
└── frontend/
    ├── css.md          # ID "css" (末尾名) または "frontend/css" (パス) で対応
    └── images/
        └── box.png
```

## ノードとの対応規則

ノードに対応する記事は次の優先順で探されます。

1. ノードに `content: <path>` があれば、そのパス (拡張子なし) の記事
2. `content/<id>.md` (ID がスラッシュ区切りのパスなら `content/<id>.md` そのもの)
3. ファイル名の末尾だけで一致する記事 (`content/frontend/css.md` → ID `css`)

```yaml
- id: css
  title: CSS
  content: frontend/css      # content/frontend/css.md を明示的に参照
```

## 同名ファイルがある場合

複数のサブディレクトリに同じ名前のファイル (例: `a/intro.md` と `b/intro.md`) があると、
末尾名での対応は **曖昧なため無効** になります。その場合は `content:` で明示するか、
ID を `a/intro` のようなパス形式にしてください。
ルート直下に同名ファイルがあるときは、ルート直下が優先されます。
`.` で始まる隠しディレクトリはスキャンされません。

## 見つからないノードの検出

対応する記事がないノードは `build` / `validate` で warning になります。
CI で止めたい場合は `roadmapper validate --strict` を使います。

## サブタスク

- [ ] サブディレクトリに記事を移動した
- [ ] 同名ファイルがあるノードに `content:` を指定した
- [ ] `validate` で未解決ノードの warning が出ないことを確認した
