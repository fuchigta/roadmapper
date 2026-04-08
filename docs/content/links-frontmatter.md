## frontmatter でリンクを追加

`.md` ファイル先頭の YAML frontmatter に `links:` を書くと、
サイドパネルのリンク一覧に表示されます。

```markdown
---
links:
  - { title: "MDN: HTML", url: "https://developer.mozilla.org/ja/docs/Web/HTML" }
  - { title: "HTML Living Standard", url: "https://html.spec.whatwg.org/" }
---

## 本文

ここから Markdown を書く...
```

## links の書式

| フィールド | 必須 | 説明 |
|---|---|---|
| `title` | 必須 | リンクの表示テキスト |
| `url` | 必須 | 遷移先 URL |

リンクはすべて `target="_blank" rel="noopener"` で新しいタブに開きます。

## roadmap.yml との優先関係

ノード定義にもリンクを書けますが、**`content/*.md` の frontmatter が優先** されます。

```yaml
# roadmap.yml (content/ の frontmatter があれば上書きされる)
nodes:
  - id: html
    title: HTML
    links:
      - { title: "MDN", url: "https://developer.mozilla.org" }
```

## サブタスク

- [ ] frontmatter に `links:` を追加した
- [ ] リンクがサイドパネルに表示されることを確認した
- [ ] `title` と `url` が正しく設定されている
