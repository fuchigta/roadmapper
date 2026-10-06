## validate --strict

`roadmapper validate` は、エラーと warning を区別します。

| 区分 | 例 | exit code |
|---|---|---|
| エラー | `title` / `id` の欠落、ID の重複、存在しない親、不正な `type` / `difficulty`、循環参照、不正な日付形式、`progressSync` / `panel` の不整合 | 1 |
| warning | ノードに対応する `content/*.md` がない、改版履歴の不整合 (未来の日付など) | 0 |

`--strict` を付けると、warning が 1 件でもあれば exit code 1 で終了します。

```bash
roadmapper validate --strict -c roadmap.yml
```

## CI で使う

記事の書き忘れや履歴の不整合を CI で止めたい場合に使います。

```yaml
# GitHub Actions の例
- run: roadmapper validate --strict
```

どの warning が原因かは標準エラー出力に一覧表示されます。

```
warning: 1 個のノードに対応する content/*.md が見つかりません:
  - [frontend] html
```

## サブタスク

- [ ] `validate` で warning の内容を確認した
- [ ] `validate --strict` がエラーゼロで通る状態にした
- [ ] CI に `validate --strict` を追加した
