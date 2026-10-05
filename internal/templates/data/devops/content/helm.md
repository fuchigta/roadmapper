---
links:
  - { title: "Helm ドキュメント", url: "https://helm.sh/docs/" }
---

## 学ぶこと

Helm は Kubernetes 向けのパッケージマネージャーです。複数のマニフェストをチャートにまとめ、値を差し替えて再利用できます。

## サブタスク

- [ ] チャートを `helm install` / `upgrade` できる
- [ ] `values.yaml` で設定を上書きできる
- [ ] チャートの基本構造を説明できる
- [ ] リリースのロールバックができる

## ポイント

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install my-nginx bitnami/nginx
```
