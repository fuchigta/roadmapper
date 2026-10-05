---
links:
  - { title: "Vue.js 公式ドキュメント", url: "https://ja.vuejs.org/" }
  - { title: "Vue: チュートリアル", url: "https://ja.vuejs.org/tutorial/" }
---

## 学ぶこと

Vue は HTML に近いテンプレート記法が特徴のフレームワークです。React の代替として選べます。

## サブタスク

- [ ] 単一ファイルコンポーネント (.vue) の構成を説明できる
- [ ] `ref` でリアクティブな状態を扱える
- [ ] `v-if` / `v-for` / `v-bind` / `v-on` を使える
- [ ] 公式チュートリアルを一通り終えられる

## ポイント

```vue
<script setup>
import { ref } from "vue";
const n = ref(0);
</script>
<template>
  <button @click="n++">{{ n }}</button>
</template>
```
