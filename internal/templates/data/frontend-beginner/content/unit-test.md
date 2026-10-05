---
links:
  - { title: "Vitest", url: "https://vitest.dev/" }
  - { title: "Jest", url: "https://jestjs.io/ja/" }
---

## 学ぶこと

関数やコンポーネントなど小さな単位が正しく動くかを確かめるテストです。
Vitest や Jest がよく使われます。

## サブタスク

- [ ] テストランナーを導入して実行できる
- [ ] `describe` / `it` / `expect` でテストを書ける
- [ ] 正常系と異常系の両方をテストできる
- [ ] テストが失敗したときの出力を読める

## ポイント

```js
import { expect, it } from "vitest";
import { sum } from "./sum";

it("足し算できる", () => {
  expect(sum(1, 2)).toBe(3);
});
```
