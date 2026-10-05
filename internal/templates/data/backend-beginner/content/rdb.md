---
links:
  - { title: "PostgreSQL チュートリアル", url: "https://www.postgresql.org/docs/current/tutorial.html" }
  - { title: "PostgreSQL 公式サイト", url: "https://www.postgresql.org/" }
---

## 学ぶこと

PostgreSQL は機能が豊富で広く使われているオープンソースのリレーショナル DB です。
データを**テーブルと関係**で表現する考え方を学びます。

## サブタスク

- [ ] PostgreSQL をローカルにインストール (または Docker で起動) できる
- [ ] `psql` でデータベースに接続できる
- [ ] テーブル設計 (正規化・インデックス) の基本を理解している
- [ ] トランザクションと ACID 特性を説明できる

## ポイント

```sql
CREATE TABLE users (
  id    SERIAL PRIMARY KEY,
  name  TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE
);
```
