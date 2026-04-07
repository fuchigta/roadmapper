# Go コーディング規約

## エラー処理

- エラーは必ずラップして文脈を付ける: `fmt.Errorf("〇〇 の処理に失敗: %w", err)`
- エラーメッセージは小文字始まり、句点なし (Go の慣習)
- `panic` は使わない。必ず `error` を返す

## 命名

- パッケージ名は短く、小文字のみ: `render`, `graph`, `meta`
- エクスポートする関数名は動詞+名詞: `RenderSVG`, `BuildGraph`, `LoadDir`
- ファイル名はスネークケース不要: `svg.go`, `html.go`, `markdown.go`

## embed FS

```go
// NG: Windows では \ 区切りになる
path := filepath.Join("data", templateName, "roadmap.yml")

// OK: embed FS は常に / 区切り
path := path.Join("data", templateName, "roadmap.yml")
```

`path` パッケージ (`path/filepath` ではなく) を embed FS 操作に使う。

## テンプレート変数

- `html/template` を使う (XSS 対策)
- 生の JS/HTML を埋め込む場合は `template.JS` / `template.HTML` を明示的に使う
- ユーザー入力を URL に埋め込む場合は `template.URL` を使う

## 定数とイミュータブル

- 設定デフォルト値は `config/loader.go` の `applyDefaults` に集約する
- 同じリテラル文字列が 3 箇所以上出てきたら定数化を検討する
