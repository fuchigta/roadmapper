---
links:
  - { title: "Go 公式インストールページ", url: "https://go.dev/dl/" }
  - { title: "roadmapper GitHub リポジトリ", url: "https://github.com/fuchigta/roadmapper" }
  - { title: "roadmapper Releases (ビルド済みバイナリ)", url: "https://github.com/fuchigta/roadmapper/releases" }
---

## 概要

roadmapper は **単一バイナリ** で動作する Go 製の CLI です。
Node.js・graphviz・Python などの外部ランタイムは一切不要です。

## 方法 1: go install

[go.dev/dl](https://go.dev/dl/) から OS に合ったインストーラをダウンロードします。
ビルドには `go.mod` に書かれた Go バージョン (Go 1.25 以上) が必要です。

```bash
# インストール確認
go version
```

```bash
go install github.com/fuchigta/roadmapper/cmd/roadmapper@latest
```

インストール後は `roadmapper` コマンドが使えるようになります。

```bash
roadmapper --help
```

## 方法 2: ビルド済みバイナリ

Go を入れたくない場合は [Releases](https://github.com/fuchigta/roadmapper/releases) から
`roadmapper-<os>-<arch>.tar.gz` (Windows は `.zip`) をダウンロードし、
展開した `roadmapper` を PATH の通った場所に置きます。

| OS | アーキテクチャ |
|---|---|
| linux / darwin | amd64 / arm64 |
| windows | amd64 |

## サブタスク

- [ ] `go install` またはビルド済みバイナリで roadmapper を入れた
- [ ] `roadmapper --help` が表示される
- [ ] `roadmapper --version` でバージョンが表示される
