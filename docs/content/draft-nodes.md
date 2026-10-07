## 下書きノード (draft)

まだ公開したくないノードは、`draft: true` で下書きにできます。指定方法は 2 つで、どちらかが `true` なら下書きです。

```yaml
# content/css.md の frontmatter
---
draft: true
---
```

```yaml
# roadmap.yml のノード (記事をまだ作っていないノード向け)
      - id: css
        title: CSS
        draft: true
```

## 公開ビルドでの表示

`roadmapper build` (既定) では、下書きノードは非活性の「準備中」ノードとして残ります。ノードとエッジの位置は変わらず、クリックしてもパネルは開きません。

- 本文・リンク・編集パスは HTML のどこにも出力されません (タイトルと親子関係だけが残ります)
- 進捗率 (ロードマップページ・トップページのカード) の分母・分子から除外されます。ブラウザに保存済みの進捗は削除されません
- 全文検索の対象外です。他ノードの関連ノード一覧には「準備中」とリンクなしで表示されます
- 下書きの `updated` / `changes` は改版履歴に集約されず、更新バッジ・RSS にも出ません。roadmap.yml の `changelog` の `nodes` に下書きノードがあれば、その参照だけ外します (項目自体は残ります)

ビルド終了時に「下書き N 件 (非公開)」と ID が表示されます。

## プレビュー

`roadmapper dev` と `roadmapper build --drafts` は、下書きを通常ノードとして出力し「下書き」バッジを付けます。公開前の確認やステージング用です。本番には使わないでください。

```bash
roadmapper build --drafts -c roadmap.yml -o dist-staging
```

## 検証

| 内容 | 扱い |
|---|---|
| 記事が未作成の下書きノード | 警告しない |
| `changelog` の `nodes` が下書きノードだけを参照している項目 | warning (`validate --strict` なら失敗) |

## サブタスク

- [ ] ノードか記事の frontmatter に `draft: true` を付けた
- [ ] `roadmapper build` で「準備中」の非活性ノードになり、本文が `dist/` に出ていないことを確認した
- [ ] `roadmapper dev` か `build --drafts` で「下書き」バッジ付きで内容を確認できた
- [ ] 公開するときに `draft: true` を外した
