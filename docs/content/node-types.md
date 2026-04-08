## ノードタイプ

ノードタイプはロードマップ上の色とバッジで視覚的に区別されます。

| タイプ | 意味 | 表示 |
|---|---|---|
| `required` | 必須 — 必ず学ぶべき項目 | 通常色 |
| `optional` | 任意 — 余裕があれば学ぶ | グレーアウト |
| `alternative` | 代替 — どれか一つ選べばよい | 点線枠 |

省略した場合は `required` と同じ扱いになります。

```yaml
nodes:
  - id: react
    title: React
    type: required       # 必須

  - id: testing
    title: テスト
    type: optional       # 任意

  - id: vue
    title: Vue
    type: alternative    # React の代替
```

## サブタスク

- [ ] 必須ノードに `type: required` を設定した
- [ ] 任意ノードに `type: optional` を設定した
- [ ] 代替ノードに `type: alternative` を設定した
