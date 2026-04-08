---
links:
  - { title: "chroma 対応言語一覧", url: "https://github.com/alecthomas/chroma#supported-languages" }
---

## コードブロックの書き方

フェンスの後ろに言語名を指定するとシンタックスハイライトが有効になります。

````markdown
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, roadmapper!")
}
```
````

## 主な対応言語

| 言語 | フェンス指定 |
|---|---|
| Go | `go` |
| JavaScript / TypeScript | `js` / `ts` |
| Python | `python` |
| Bash / Shell | `bash` / `sh` |
| YAML | `yaml` |
| JSON | `json` |
| HTML | `html` |
| CSS | `css` |
| SQL | `sql` |

200 以上の言語に対応しています。不明な場合は [chroma 対応言語一覧](https://github.com/alecthomas/chroma#supported-languages) を参照してください。

## テーマ

ライトモード / ダークモード それぞれに最適化されたハイライトが自動適用されます。
テーマはサイトのダークモード切替と連動します。

## サブタスク

- [ ] コードブロックに言語名を指定した
- [ ] ハイライトが正しく表示されることを `dev` で確認した
