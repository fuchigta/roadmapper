## 親子関係の定義

ノードの依存関係は `children` と `parents` で表現します。

```yaml
nodes:
  - id: html
    title: HTML
    children:
      - id: html-basics
        title: 基本要素

  - id: css
    title: CSS
    parents: [html]      # html を学んだ後に CSS
    children:
      - id: css-basics
        title: セレクタ・ボックスモデル
```

## 複数親 (DAG)

1 つのノードが複数の親を持つことができます。

```yaml
  - id: framework
    title: フレームワーク
    parents: [javascript, css]   # JS と CSS を学んだ後
```

これにより単純なツリーではなく DAG (有向非巡回グラフ) を表現できます。

> **循環参照は禁止** — A → B → A のような循環があると `validate` がエラーを返します。

## サブタスク

- [ ] `children` でノードの階層を定義した
- [ ] `parents` で複数の依存関係を表現した
- [ ] `validate` で循環参照がないことを確認した
