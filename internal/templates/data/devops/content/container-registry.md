---
links:
  - { title: "Docker Hub ドキュメント", url: "https://docs.docker.com/docker-hub/" }
---

## 学ぶこと

コンテナレジストリはビルドしたイメージを保管・配布する場所です。CI/CD からのプッシュと、デプロイ時のプルで使います。

## サブタスク

- [ ] タグとダイジェストの違いを説明できる
- [ ] イメージにタグを付けてプッシュできる
- [ ] プライベートレジストリの認証方法を理解している
- [ ] `latest` タグに頼らない理由を説明できる

## ポイント

```bash
docker tag myapp:1.0 registry.example.com/team/myapp:1.0
docker push registry.example.com/team/myapp:1.0
```
