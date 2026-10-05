---
links:
  - { title: "Docker ドキュメント", url: "https://docs.docker.com/" }
---

## 学ぶこと

Docker はコンテナを作成・実行するための代表的なツールです。

## サブタスク

- [ ] `docker run` でコンテナを起動できる
- [ ] `docker ps` / `logs` / `exec` で状態を確認できる
- [ ] ポート公開とボリュームを使える
- [ ] 不要なコンテナやイメージを削除できる

## ポイント

```bash
docker run --rm -d -p 8080:80 nginx
docker ps
docker logs <container-id>
```
