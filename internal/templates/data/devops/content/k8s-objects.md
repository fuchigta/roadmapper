---
links:
  - { title: "Kubernetes: ワークロード", url: "https://kubernetes.io/ja/docs/concepts/workloads/" }
  - { title: "Kubernetes: Service", url: "https://kubernetes.io/ja/docs/concepts/services-networking/service/" }
---

## 学ぶこと

Kubernetes の基本リソースである Pod / Deployment / Service を理解しましょう。

## サブタスク

- [ ] Pod が最小のデプロイ単位であることを説明できる
- [ ] Deployment でレプリカ数とローリングアップデートを管理できる
- [ ] Service でクラスター内外に公開できる
- [ ] ConfigMap / Secret の使い分けを説明できる

## ポイント

```bash
kubectl create deployment web --image=nginx --replicas=2
kubectl expose deployment web --port=80
```
