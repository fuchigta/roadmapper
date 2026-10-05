---
links:
  - { title: "Docker: Get started", url: "https://docs.docker.com/get-started/" }
  - { title: "Dockerfile リファレンス", url: "https://docs.docker.com/reference/dockerfile/" }
---

## 学ぶこと

Docker は、アプリと実行環境を**コンテナ**にまとめて、どこでも同じように動かす技術です。
「自分の環境では動いたのに」という問題を減らせます。

## サブタスク

- [ ] イメージとコンテナの違いを説明できる
- [ ] `docker run` でコンテナを起動・停止できる
- [ ] `Dockerfile` で自分のアプリのイメージをビルドできる
- [ ] ポートやボリュームをコンテナにマッピングできる
- [ ] Docker Compose で複数コンテナを起動できる

## ポイント

```dockerfile
FROM golang:1.22
WORKDIR /app
COPY . .
RUN go build -o server .
CMD ["./server"]
```
