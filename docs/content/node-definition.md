## ノードの基本構造

1 つのノードは以下のキーで構成されます。

```yaml
nodes:
  - id: html-basics            # ノードの一意 ID (必須)
    title: HTML の基礎          # サイト上に表示される名前 (必須)
    type: required             # 省略可 (省略時は required)
    difficulty: beginner       # 難易度 (省略可): beginner / intermediate / advanced
    estimatedTime: 3d          # 推定所要時間 (省略可): 自由書式 (例: "30m", "2h", "3d", "2w")
    content: frontend/html     # 対応する content/<path>.md を明示 (省略可、拡張子なし)
    links:                     # 参考リンク (省略可。content の frontmatter があればそちらが優先)
      - { title: MDN, url: "https://developer.mozilla.org/" }
    x: 300                     # 手動座標 (省略可。自動レイアウトを上書き)
    y: 200
    children:                  # インライン定義の子ノード
      - id: semantic-html
        title: セマンティック HTML
```

`id` と `title` は必須です。空だと `validate` / `build` がエラーになります。

## difficulty (難易度)

ノードの難易度を表す任意フィールドです。設定すると SVG ノードの左上にバッジ (初 / 中 / 上) が表示され、サイドパネルにも難易度ラベルが表示されます。
上記以外の値を書くと `validate` がエラーにします。

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

## x / y (手動座標)

通常はレイアウトが自動で座標を決めますが、`x` / `y` を指定するとそのノードだけ位置を上書きできます。
片方だけの指定も可能です。

## id の命名ルール

- **英小文字・数字・ハイフン** を使う (`html-basics`, `step-1`)
- **同一ロードマップ内で一意** であること (ロードマップをまたいで重複しても可。ただし `__order` は予約済みで使えない)
- 既定では `content/<id>.md` と対応するため、**ファイル名に使えない文字は避ける**
  (`frontend/html` のようなスラッシュ区切りの ID を使うと `content/frontend/html.md` に対応する。詳しくは「サブディレクトリとノード ID の対応」)

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
