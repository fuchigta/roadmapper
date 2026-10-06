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

ファイルが存在しないノードはサイドパネルが空欄になり、`build` / `validate` で warning が表示されます。
`content/` はサブディレクトリに分けて整理することもできます (下記「サブディレクトリとノード ID の対応」)。

## frontmatter

`.md` の先頭に YAML frontmatter を書けます。すべて省略可能です。

| キー | 説明 |
|---|---|
| `title` | 任意。ノードの表示名は `roadmap.yml` の `title` が正 |
| `links` | 参考リンク。`roadmap.yml` のノードの `links` より優先される |
| `updated` | ノードの最終更新日 (`YYYY-MM-DD`) |
| `changes` | ノード単位の改版履歴 (`date` / `summary` のリスト) |

## 学ぶこと

- **frontmatter でリンクを追加** — `links:` キーで参考 URL をサイドパネルに表示する方法
- **チェックリストで進捗管理** — GFM タスクリスト構文と localStorage への保存
- **コードブロックとシンタックスハイライト** — 言語指定フェンスと chroma テーマ
- **Mermaid 図の埋め込み** — フローチャートやシーケンス図を Markdown 内に記述する方法
- **サブディレクトリとノード ID の対応** — `content/` の整理と `content:` による明示指定
- **画像など静的ファイルの参照** — 相対パスでの参照と `contentAssets.exclude`
- **改版履歴 (changelog)** — roadmap.yml と frontmatter で更新履歴を管理する方法
