---
links:
  - { title: "React 公式ドキュメント", url: "https://ja.react.dev/" }
  - { title: "React: クイックスタート", url: "https://ja.react.dev/learn" }
---

## 学ぶこと

React はコンポーネントで UI を組み立てるライブラリです。国内外で広く使われています。

## サブタスク

- [ ] 関数コンポーネントと JSX を書ける
- [ ] props でコンポーネントにデータを渡せる
- [ ] `useState` で状態を管理できる
- [ ] `useEffect` の役割を説明できる
- [ ] リストのレンダリングで `key` を指定できる

## ポイント

```jsx
function Counter() {
  const [n, setN] = useState(0);
  return <button onClick={() => setN(n + 1)}>{n}</button>;
}
```
