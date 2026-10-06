## 画像など静的ファイルの参照

`content/` 配下に画像や PDF などの `.md` 以外のファイルを置くと、ビルド時に
`dist/content/` へ同じ構造でコピーされます。

```
content/
└── frontend/
    ├── html.md
    └── images/
        └── dom.png
```

Markdown からは、その記事 (`.md`) からの **相対パス** で参照します。

```markdown
<!-- content/frontend/html.md 内 -->
![DOM の図](./images/dom.png)
```

ビルド時に URL はロードマップページから正しく解決できるパス (例: `../content/frontend/images/dom.png`) に
自動で書き換えられます。`basePath` を設定している場合も同様です。

絶対 URL (`https://…`)、ルート相対パス (`/foo`)、`mailto:`、フラグメント (`#section`) は
そのまま保持されます。

## コピー対象から除外する

`site.contentAssets.exclude` に glob パターンを書くと、コピーしないファイルを指定できます。

```yaml
site:
  contentAssets:
    exclude:
      - "**/drafts/**"
      - "*.psd"
      - "**/raw/*"
```

| パターン | 意味 |
|---|---|
| `*` | 1 つのパスセグメント内の任意の文字列 |
| `**` | 0 個以上のセグメント |

`*.psd` は `content/` 直下のファイルにだけ一致します。サブディレクトリ内も含めて除外するには `**/*.psd` と書きます。
パターンは `content/` からの相対パスに対して評価されます。`.` で始まる隠しディレクトリは常に対象外です。

## サブタスク

- [ ] 画像を `content/` 配下に置き、相対パスで参照した
- [ ] ビルド後に `dist/content/` へコピーされたことを確認した
- [ ] 公開したくないファイルを `contentAssets.exclude` に追加した
