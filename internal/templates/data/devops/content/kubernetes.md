---
links:
  - { title: "Kubernetes ドキュメント (日本語)", url: "https://kubernetes.io/ja/docs/home/" }
  - { title: "kubectl リファレンス", url: "https://kubernetes.io/docs/reference/kubectl/" }
---

## 学ぶこと

Kubernetes はコンテナオーケストレーションのデファクトスタンダードです。

## サブタスク

- [ ] コントロールプレーンとノードの役割を説明できる
- [ ] ローカル環境 (kind / minikube など) でクラスターを用意できる
- [ ] `kubectl` で状態の確認ができる
- [ ] マニフェストを `kubectl apply` で適用できる

## ポイント

```bash
kubectl get nodes
kubectl get pods -A
kubectl apply -f deployment.yaml
```
