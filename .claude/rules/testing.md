# テスト方針

## 基本ルール

- `go test ./...` が常に通ること
- 新しい `internal/` パッケージを作る場合はテストファイルも作る
- `internal/command/` は統合テストが難しいため手動確認で可

## テストの書き方

```go
// テーブル駆動テストを優先する
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"正常ケース", "input", "want", false},
        {"エラーケース", "bad", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Foo(tt.input)
            if (err != nil) != tt.wantErr {
                t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

## ゴールデンファイルテスト

SVG や HTML の出力確認には `testdata/` のゴールデンファイルを使う:

```go
// 初回: UPDATE_GOLDEN=1 go test ./... でファイル生成
// 以降: go test ./... で比較
if os.Getenv("UPDATE_GOLDEN") == "1" {
    os.WriteFile("testdata/want.html", []byte(got), 0644)
}
want, _ := os.ReadFile("testdata/want.html")
if got != string(want) {
    t.Errorf("output mismatch:\ngot:  %s\nwant: %s", got, want)
}
```

## モックについて

- 外部 I/O (ファイル, ネットワーク) はインターフェースで抽象化してモック可能にする
- ただし過度なモック化は避ける。実際のファイル操作は `t.TempDir()` を使えば十分クリーンアップされる

```go
func TestLoadDir(t *testing.T) {
    dir := t.TempDir()
    os.WriteFile(filepath.Join(dir, "foo.md"), []byte("# Foo"), 0644)
    docs, err := LoadDir(dir)
    // ...
}
```
