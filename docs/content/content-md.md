---
links:
  - { title: "goldmark (Markdown パーサ)", url: "https://github.com/yuin/goldmark" }
  - { title: "chroma (シンタックスハイライト)", url: "https://github.com/alecthomas/chroma" }
---

## コンテンツ Markdown の配置

ノード ID と同じ名前の `.md` ファイルを `content/` ディレクトリに置くと、
サイドパネルに詳細説明が表示されます。

```
content/
├── html.md        # id: html のノードに対応
├── css.md         # id: css のノードに対応
└── javascript.md
```

## frontmatter でリンクを追加

ファイル先頭の YAML frontmatter でリンクリストを定義できます。

```markdown
---
links:
  - { title: "MDN: HTML", url: "https://developer.mozilla.org/ja/docs/Web/HTML" }
  - { title: "HTML Living Standard", url: "https://html.spec.whatwg.org/" }
---

## 学ぶこと

本文をここに書く...
```

## チェックリストで進捗管理

GFM のタスクリスト構文を使うと、ノードの進捗チェックボックスが有効になります。

```markdown
## サブタスク

- [ ] `<header>` / `<main>` / `<footer>` を使い分けられる
- [ ] フォームを正しく書ける
- [x] インストールが完了した  ← チェック済みの例
```

チェックした状態は localStorage に保存されます。

## サブタスク

- [ ] ノード ID と同じ名前の `.md` ファイルを `content/` に作成した
- [ ] frontmatter でリンクを追加した
- [ ] チェックリストを書いて進捗管理を試した
