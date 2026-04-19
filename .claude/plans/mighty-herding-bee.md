# 進捗インポート/エクスポート削除 + ノード全文検索

## Context

現在の roadmapper には2つの問題がある。

1. **使われないインポート/エクスポート機能が UI にある**: `web/templates/roadmap.html` のツールバーに「エクスポート」「インポート」ボタンがあり、localStorage の JSON ダンプをダウンロード/アップロードする。URL でシェアする `share-btn` が別途あり、実質的にこちらで用が足りる。使い道が薄いので UI から落としたい。
2. **検索がタイトル一致のみで弱い**: `initSearch` は `nodeData[id].title` に対する `.toLowerCase().includes(q)` だけ。ノード本文 (Markdown) や関連リンクタイトルに出てくる語句は拾えず、ロードマップが成長するほど検索の有用性が下がる。

本プランでは (A) インポート/エクスポート機能を完全削除し、(B) ビルド時に各ノードの plaintext をサーバ側 (Go) で生成して `ROADMAP_DATA` に同梱、クライアントはタイトル + 本文 + リンクタイトルにまたがる全文検索をスニペット+ハイライト付きで提供する。共有は `share-btn` (URL パラメータ) に一本化する。

## 設計方針

- **外部ランタイムゼロ**の原則は維持。テキスト抽出は Go 側で goldmark AST を walk して実装 (既存の goldmark v1.8.2 を流用、新規依存なし)。
- **ペイロード**: `docs/` 最大ロードマップでも ROADMAP_DATA 67KB。plaintext 追加で +20-30KB 程度、許容範囲。別ファイルには分けない。
- **JS は素の JS のまま** (目標 30KB 以内)。ライブラリは追加しない。
- **日本語**: substring 検索 (`includes`) で済ませる。形態素解析は入れない。

## 変更ファイル

### Part 1: インポート/エクスポート削除

#### `web/templates/roadmap.html`
- **L45-46, L48** を削除 (`#export-btn`, `#import-btn`, `#import-file`)。L47 の `#share-btn` は残す。`.toolbar` のレイアウトは search-wrap + share-btn のみになる。

#### `web/static/app.js`
- **L660-690** のエクスポート/インポートハンドラ2ブロックを削除。他に参照なし。
- `DOMContentLoaded` 内のコメント「// エクスポート」「// インポート」も一緒に除去。

#### `README.md`
- **L11** の機能リストから「JSON エクスポート/インポート」の記述を削除。
- **L211-L212** の操作説明テーブル2行を削除。

#### CSS
- `web/static/style.css` には `#export-btn` / `#import-btn` 専用ルール無し (generic `.toolbar button` 継承のみ)。修正不要。

### Part 2: 全文検索

#### `internal/render/markdown.go` — plaintext 抽出関数を追加

新規エクスポート関数:

```go
// ExtractPlainText は Markdown ソースからプレーンテキストを抽出する。
// ヘッダ/本文/リストテキスト + コードブロック内容を空白連結して返す。
// 全文検索インデックス用。HTML タグやエンティティは出現しない。
func ExtractPlainText(src string) string
```

実装: `md.Parser().Parse(text.NewReader([]byte(src)))` で AST を取り、`ast.Walk` で以下を拾う。
- `*ast.Text`: `v.Segment.Value(source)` を書き出し
- `*ast.FencedCodeBlock` / `*ast.CodeBlock`: `Lines()` をイテレートして `seg.Value(source)` を書き出し
- `*ast.AutoLink`: `v.URL(source)` / `v.Label(source)`
- それ以外はスキップ (継続)

最後に `strings.Join(strings.Fields(...), " ")` で空白正規化。

goldmark の import が必要 (`github.com/yuin/goldmark/ast`, `github.com/yuin/goldmark/text`) — 既存 go.mod に goldmark は入っている。

#### `internal/render/html.go` — NodeMeta 拡張

- `NodeMeta` 構造体に `Text string \`json:"text,omitempty"\`` を追加 (L16-25)。
- `buildNodeMeta` の引数に `nodeText map[string]string` を追加し、`Text: nodeText[n.ID]` を埋める。
- `RenderRoadmapPage` の引数に `nodeText map[string]string` を追加し `buildNodeMeta` へ引き渡す。

#### `internal/command/build.go` — plaintext 生成

`buildNodeHTML` を拡張するか、並行して `buildNodeText` を作るか。**拡張案**が簡潔:

- 関数名を `buildNodeAssets` にリネーム (または戻り値を増やす)。返り値: `(nodeHTML map[string]string, nodeText map[string]string, hasMermaid bool, err error)`。
- ループ内で:
  - 本文 plaintext: `textBody := render.ExtractPlainText(doc.Body)`
  - リンクタイトル連結: `linkText := strings.Join(titles, " ")` (`n.Node.Links` からタイトルを拾う。content frontmatter 優先マージ後の値を使う)
  - `nodeText[n.ID] = strings.TrimSpace(textBody + " " + linkText)`
- `runBuild` の呼び出し側で `nodeText` を `render.RenderRoadmapPage` に渡す。

#### `web/static/app.js` — initSearch 書き換え (L324-403)

現在の実装を以下に置き換える。

**検索インデックス構築** (`initSearch` 内で一度だけ):
```js
const idx = Object.entries(nodeData)
  .filter(([id]) => id !== '__order')
  .map(([id, d]) => ({
    id,
    title: d.title,
    titleLower: d.title.toLowerCase(),
    text: d.text || '',
    textLower: (d.text || '').toLowerCase(),
  }));
```

**マッチング** (`showResults(query)` 内):
- `q = query.trim().toLowerCase()`; 空なら非表示。
- 各エントリで:
  - `titleIdx = entry.titleLower.indexOf(q)`
  - `textIdx = entry.textLower.indexOf(q)`
  - いずれもヒットしない → スキップ
  - `score = titleIdx >= 0 ? (titleIdx === 0 ? 0 : 1) : 2` (先頭マッチ < 部分マッチ < 本文のみ)
  - `snippet`: 本文マッチあれば `entry.text.slice(max(0, textIdx-30), textIdx+q.length+50)` に前後省略記号を付ける。本文マッチなし (タイトルのみヒット) の場合は snippet 省略。
- `score` 昇順で先頭 12 件を表示。

**DOM 構築**:
- 各 `<li>` を textContent ではなく DOM で組む:
  - `<div class="sr-title">` にタイトルを `<mark>` 区切りで挿入 (小文字インデックスを使って元文字列をスライスし、`document.createElement('mark')` で XSS 安全に)。
  - 本文マッチあれば `<div class="sr-snippet">` を追加し、同様にハイライト。
- `li.dataset.id = entry.id` に格納し、クリックは dataset 経由で `selectResult` を呼ぶ。

**Enter キーのバグ修正** (L382-385):
現状は `nodeData` を再フィルタして `selectedIdx` 番目を取っているが、マッチング基準がずれると誤ったノードを開く。マッチ配列 (`lastMatches`) を closure に保持し、`items[selectedIdx].dataset.id` を使って直接解決する。

**ハイライトヘルパー**:
```js
function appendWithMark(parent, text, lowerText, q) {
  let i = 0;
  while (true) {
    const hit = lowerText.indexOf(q, i);
    if (hit < 0) { parent.appendChild(document.createTextNode(text.slice(i))); return; }
    parent.appendChild(document.createTextNode(text.slice(i, hit)));
    const m = document.createElement('mark');
    m.textContent = text.slice(hit, hit + q.length);
    parent.appendChild(m);
    i = hit + q.length;
  }
}
```

#### `web/static/style.css` — 検索結果のスタイル追加

新規セレクタ (既存のカラー変数を流用):
- `.search-results li .sr-title` — 通常の文字色
- `.search-results li .sr-snippet` — `font-size: 0.85em; color: var(--text-muted); margin-top: 2px;` 風
- `.search-results mark` — ブランドカラーで背景、`border-radius: 2px; padding: 0 2px;`

#### テスト

- `internal/render/markdown_test.go` に `TestExtractPlainText` を追加 (テーブル駆動)。ケース: 見出し+段落, コードフェンス, リスト, 太字/リンク (リンク文字は残す), HTML タグ混在。
- クライアント側は手動確認 (demo または docs で build → `dev` で確認)。

## 検証手順

```bash
# 1. ビルドと単体テスト (lefthook pre-commit 相当)
go build ./...
go test ./...

# 2. デモで動作確認
go run ./cmd/roadmapper build -c docs/roadmap.yml -o docs/dist
go run ./cmd/roadmapper dev -c docs/roadmap.yml   # http://localhost:3000/
```

ブラウザで:
- `/` キーで検索フォーカス → 本文にしかない語 (例: docs のコード内の `go build` など) を入力 → ヒットしてスニペットに `<mark>` が出る。
- タイトル先頭一致が本文のみヒットより上位に表示される。
- 結果を Enter / クリック で開くとそのノードのパネルが表示される。
- ツールバーから「エクスポート」「インポート」ボタンが消え、「シェア」のみ残ること。
- シェア URL の読み書きが従来通り動くこと。

## 変更しないもの

- `share-btn` と `encodeProgress`/`decodeProgress` — 独立機能なので残す。
- `.toolbar button` の汎用 CSS — shareBtn が引き続き使う。
- `ROADMAP_DATA` の既存フィールド (title/html/parents/children/…) — 互換維持、`text` だけ追加。
- `.claude/plans/cheerful-wibbling-snowflake.md` などの歴史的ドキュメント — 触らない。
