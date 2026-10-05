---
links:
  - { title: "Prometheus ドキュメント", url: "https://prometheus.io/docs/" }
  - { title: "Grafana ドキュメント", url: "https://grafana.com/docs/" }
---

## 学ぶこと

Prometheus はメトリクス収集・保存のための OSS、Grafana はそれを可視化するツールです。

## サブタスク

- [ ] Pull 型のメトリクス収集を説明できる
- [ ] Counter / Gauge / Histogram の違いを説明できる
- [ ] PromQL で簡単なクエリを書ける
- [ ] Grafana でダッシュボードを作れる

## ポイント

```text
# 直近 5 分の HTTP リクエスト毎秒 (PromQL)
rate(http_requests_total[5m])
```
