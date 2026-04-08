## 複数ロードマップの定義

`roadmaps:` は配列なので、1 つの `roadmap.yml` に複数のロードマップを定義できます。

```yaml
roadmaps:
  - id: frontend
    title: フロントエンド
    description: HTML・CSS・JavaScript の基礎から応用まで
    nodes:
      - id: html
        title: HTML

  - id: backend
    title: バックエンド
    description: サーバサイドの基礎
    nodes:
      - id: server-basics
        title: サーバの基礎
```

生成されるサイト構造:

```
dist/
├── index.html          # ロードマップ一覧 (両方のカードが表示される)
├── frontend/
│   └── index.html
└── backend/
    └── index.html
```

## ロードマップ間のルール

- **ロードマップ `id` はサイト内で一意** であること
- ノード `id` は異なるロードマップ間では重複しても構わない
- ロードマップをまたいだ `parents` / `children` 参照は **不可**

## 使い分けの目安

| ユースケース | 推奨 |
|---|---|
| 段階的な学習パス (初級→中級→上級) | 複数ロードマップ |
| 独立した技術領域 (フロント / バック / インフラ) | 複数ロードマップ |
| 1 つのスキルを深く掘り下げる | 単一ロードマップで親子階層を使う |

## サブタスク

- [ ] 複数の `roadmaps:` エントリを定義した
- [ ] 各ロードマップに `id`, `title`, `description` を設定した
- [ ] `validate` でエラーゼロを確認した
- [ ] インデックスページで複数カードが表示されることを確認した
