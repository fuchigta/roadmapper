---
links:
  - { title: "MDN: JSON の操作", url: "https://developer.mozilla.org/ja/docs/Learn/JavaScript/Objects/JSON" }
  - { title: "JSON 公式サイト", url: "https://www.json.org/json-ja.html" }
---

## 学ぶこと

JSON は API でデータをやり取りする際の標準的な形式です。プログラム内のデータを JSON 文字列に変換する
**シリアライズ**と、その逆の**デシリアライズ**を学びます。

## サブタスク

- [ ] JSON の文法 (オブジェクト・配列・文字列・数値・真偽値・null) を理解している
- [ ] 選んだ言語で JSON のエンコード・デコードができる
- [ ] 日時や欠損値の扱いなど、型の違いに注意すべき点を知っている

## ポイント

```json
{
  "id": 1,
  "name": "Taro",
  "tags": ["admin", "dev"],
  "deletedAt": null
}
```

> JSON にコメントは書けず、末尾のカンマも許されません。
