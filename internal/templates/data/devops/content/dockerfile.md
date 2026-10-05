---
links:
  - { title: "Dockerfile リファレンス", url: "https://docs.docker.com/reference/dockerfile/" }
---

## 学ぶこと

Dockerfile はイメージの作り方を記述するファイルです。小さく安全なイメージを作る工夫を学びましょう。

## サブタスク

- [ ] `FROM` / `COPY` / `RUN` / `CMD` を使って書ける
- [ ] レイヤーキャッシュを意識して命令を並べられる
- [ ] マルチステージビルドを説明できる
- [ ] root 以外のユーザーで実行する理由を理解している

## ポイント

```dockerfile
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /app .

FROM gcr.io/distroless/static
COPY --from=build /app /app
CMD ["/app"]
```
