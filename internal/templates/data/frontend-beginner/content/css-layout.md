---
links:
  - { title: "MDN: Flexbox の基本概念", url: "https://developer.mozilla.org/ja/docs/Web/CSS/CSS_flexible_box_layout/Basic_concepts_of_flexbox" }
  - { title: "MDN: CSS グリッドレイアウト", url: "https://developer.mozilla.org/ja/docs/Web/CSS/CSS_grid_layout" }
---

## 学ぶこと

現代の CSS レイアウトは **Flexbox** (1 次元) と **Grid** (2 次元) が中心です。

## サブタスク

- [ ] `display: flex` で横並び・中央揃えができる
- [ ] `justify-content` / `align-items` の違いを説明できる
- [ ] `display: grid` と `grid-template-columns` で段組みを作れる
- [ ] Flexbox と Grid の使い分けを説明できる

## ポイント

```css
.row { display: flex; gap: 8px; }
.grid { display: grid; grid-template-columns: repeat(3, 1fr); }
```
