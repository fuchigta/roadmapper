---
links:
  - { title: "PostgreSQL チュートリアル: SQL 言語", url: "https://www.postgresql.org/docs/current/tutorial-sql.html" }
---

## 学ぶこと

SQL はリレーショナル DB を操作するための言語です。データの取得・追加・更新・削除と、
複数テーブルの結合を使いこなせるようにしましょう。

## サブタスク

- [ ] `SELECT` で条件指定・並び替え・件数制限ができる
- [ ] `INSERT` / `UPDATE` / `DELETE` でデータを変更できる
- [ ] `JOIN` で複数テーブルを結合できる
- [ ] `GROUP BY` と集計関数を使える

## ポイント

```sql
SELECT u.name, COUNT(o.id) AS order_count
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
GROUP BY u.name;
```

> `UPDATE` / `DELETE` には必ず `WHERE` を付けましょう。
