## ノードの基本構造

1 つのノードは以下のキーで構成されます。

```yaml
nodes:
  - id: html-basics        # ノードの一意 ID (英小文字・数字・ハイフン推奨)
    title: HTML の基礎      # サイト上に表示される名前
    type: required         # 省略可 (省略時は required と同等)
    children:              # インライン定義の子ノード
      - id: semantic-html
        title: セマンティック HTML
```

## id の命名ルール

- **英小文字・数字・ハイフン** のみ使用する (`html-basics`, `step-1`)
- **同一ファイル内で一意** であること (ロードマップをまたいで重複しても可)
- `content/<id>.md` と 1:1 で対応するため、**ファイル名に使えない文字は避ける**

## 子ノードの定義方法

`children:` に直接インライン定義する方法と、`parents:` で後から参照する方法の 2 通りがあります。

```yaml
# インライン定義
nodes:
  - id: css
    title: CSS
    children:
      - id: flexbox
        title: Flexbox

# parents 参照 (同階層で定義)
  - id: flexbox
    title: Flexbox
    parents: [css]
```

## サブタスク

- [ ] 全ノードに一意の `id` を設定した
- [ ] `id` に英小文字・ハイフンのみ使用している
- [ ] `title` が学習者にわかりやすい名前になっている
- [ ] `validate` でエラーゼロを確認した
