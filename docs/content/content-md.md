---
links:
  - { title: "goldmark (Markdown パーサ)", url: "https://github.com/yuin/goldmark" }
  - { title: "chroma (シンタックスハイライト)", url: "https://github.com/alecthomas/chroma" }
---

## コンテンツ Markdown の配置

ノード ID と同じ名前の `.md` ファイルを `content/` ディレクトリに置くと、
クリック時のサイドパネルに詳細説明が表示されます。

```
content/
├── html.md        # id: html のノードに対応
├── css.md         # id: css のノードに対応
└── javascript.md
```

ファイルが存在しないノードはサイドパネルが空欄になります。

## 学ぶこと

- **frontmatter でリンクを追加** — `links:` キーで参考 URL をサイドパネルに表示する方法
- **チェックリストで進捗管理** — GFM タスクリスト構文と localStorage への保存
- **コードブロックとシンタックスハイライト** — 言語指定フェンスと chroma テーマ
- **Mermaid 図の埋め込み** — フローチャートやシーケンス図を Markdown 内に記述する方法
