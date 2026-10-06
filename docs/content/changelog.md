## 改版履歴 (changelog)

ロードマップの更新履歴を 2 か所で書けます。日付はすべて `YYYY-MM-DD` 形式です。

**roadmap.yml** — ロードマップ全体に関わる変更 (構成の見直し、複数ノードにまたがる変更)

```yaml
roadmaps:
  - id: frontend
    title: Frontend
    changelog:
      - date: 2026-09-20
        summary: フォームの章を追加
        nodes: [html]        # 関連ノード ID (任意)
```

**content/*.md の frontmatter** — そのノード自身の変更

```yaml
---
updated: 2026-09-20          # 最終更新日
changes:
  - date: 2026-09-20
    summary: アクセシビリティの説明を追記
---
```

## 表示される場所

両者は 1 つの履歴 (日付降順) に集約され、次の場所に表示されます。

- ロードマップページのヘッダー (最新の更新日・履歴)
- サイドパネル (ノードごとの履歴)
- トップページのロードマップカード
- SVG の「更新」バッジ (最終更新から 30 日以内のノード。ホバーで更新日を表示)
- RSS フィード (`siteUrl` 設定時)。履歴のあるロードマップは履歴項目ごとに item を出力する

## 検証

| 内容 | 扱い |
|---|---|
| `date` が空・`YYYY-MM-DD` でない、`summary` が空、`nodes` に存在しない ID | エラー (`validate` / `build` が失敗) |
| 未来の日付、`updated` が `changes` / roadmap.yml の changelog より古い | warning (`validate --strict` なら失敗) |

## サブタスク

- [ ] roadmap.yml の `changelog` に履歴を 1 件追加した
- [ ] 記事の frontmatter に `updated` / `changes` を追加した
- [ ] ビルドして更新履歴ボタンとパネルの履歴が表示されることを確認した
- [ ] `validate --strict` で警告が出ないことを確認した
