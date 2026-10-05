---
links:
  - { title: "A Tour of Go", url: "https://go.dev/tour/" }
  - { title: "Go ドキュメント", url: "https://go.dev/doc/" }
---

## 学ぶこと

Go はシンプルな文法と高い並行処理性能を持つコンパイル言語で、API サーバーや CLI ツールで
広く使われています。標準ライブラリだけで HTTP サーバーが書けるのも魅力です。

## サブタスク

- [ ] A Tour of Go を一通り終える
- [ ] 構造体・インターフェース・エラー処理の基本を理解している
- [ ] goroutine と channel の基本を説明できる
- [ ] `net/http` で簡単な HTTP サーバーを書ける
- [ ] `go mod` でパッケージ管理ができる

## ポイント

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Go!")
	})
	http.ListenAndServe(":8080", nil)
}
```
