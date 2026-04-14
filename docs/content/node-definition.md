## ノードの基本構造

1 つのノードは以下のキーで構成されます。

```yaml
nodes:
  - id: html-basics            # ノードの一意 ID (英小文字・数字・ハイフン推奨)
    title: HTML の基礎          # サイト上に表示される名前
    type: required             # 省略可 (省略時は required と同等)
    difficulty: beginner       # 難易度 (省略可): beginner / intermediate / advanced
    estimatedTime: 3d          # 推定所要時間 (省略可): 自由書式 (例: "30m", "2h", "3d", "2w")
    children:                  # インライン定義の子ノード
      - id: semantic-html
        title: セマンティック HTML
```

## difficulty (難易度)

ノードの難易度を表す任意フィールドです。設定すると SVG ノードの左上にバッジが表示され、サイドパネルにも難易度ラベルが表示されます。

| 値 | 表示 | 意味 |
|---|---|---|
| `beginner` | 初級 (緑) | 基礎的な内容 |
| `intermediate` | 中級 (黄) | ある程度の前提知識が必要 |
| `advanced` | 上級 (赤) | 高度な内容 |

## estimatedTime (推定所要時間)

学習にかかる目安時間を自由書式の文字列で記述します。サイドパネルに表示されます。省略可能です。

```yaml
estimatedTime: "30m"   # 30分
estimatedTime: "2h"    # 2時間
estimatedTime: "3d"    # 3日
estimatedTime: "2w"    # 2週間
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
