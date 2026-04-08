---
links:
  - { title: "Go 公式インストールページ", url: "https://go.dev/dl/" }
  - { title: "roadmapper GitHub リポジトリ", url: "https://github.com/fuchigta/roadmapper" }
---

## 概要

roadmapper は **単一バイナリ** で動作する Go 製の CLI です。
Node.js・graphviz・Python などの外部ランタイムは一切不要です。

## Go のインストール

[go.dev/dl](https://go.dev/dl/) から OS に合ったインストーラをダウンロードします。

```bash
# インストール確認
go version
# go version go1.22.x ...
```

## roadmapper のインストール

```bash
go install github.com/fuchigta/roadmapper/cmd/roadmapper@latest
```

インストール後は `roadmapper` コマンドが使えるようになります。

```bash
roadmapper --help
```

## サブタスク

- [ ] Go 1.21 以上がインストールされている
- [ ] `go install` が完了している
- [ ] `roadmapper --help` が表示される
