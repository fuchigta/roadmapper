## ノードタイプ

ノードタイプはロードマップ上の色とバッジで視覚的に区別されます。

| タイプ | 意味 | 表示 |
|---|---|---|
| `required` | 必須 — 必ず学ぶべき項目 | 濃色の塗り |
| `optional` | 任意 — 余裕があれば学ぶ | 薄色の塗り + グレー枠、`opt` バッジ。親からのエッジは破線 |
| `alternative` | 代替 — どれか一つ選べばよい | 紫系の塗り + `alt` バッジ。親からのエッジは破線 |

省略した場合は `required` と同じ扱いになります。
上記以外の値を書くと `validate` がエラーにします。

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
