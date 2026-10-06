## validate — 設定ファイルのチェック

ビルド前に `roadmap.yml` の構文・必須項目・重複 ID・親ノードの存在・循環参照などを検証します。

```bash
roadmapper validate -c my-roadmap/roadmap.yml
```

エラーがなければ `✓ ... の検証が完了しました` と、解決できた content の件数が表示されます。
`-c` を省略するとカレントディレクトリの `roadmap.yml` を使います。

ノードに対応する `content/*.md` が見つからない場合や、改版履歴に不整合がある場合は
**warning** として表示されます (exit code は 0)。

## dev — ライブプレビュー

ファイルを保存するたびにブラウザが自動更新される開発サーバです。
`roadmap.yml` と `content/` (サブディレクトリを含む) を監視します。

```bash
roadmapper dev -c my-roadmap/roadmap.yml
# http://localhost:4321 で起動
```

| オプション | 既定値 | 説明 |
|---|---|---|
| `-c` / `--config` | `roadmap.yml` | 設定ファイルのパス |
| `-o` / `--out` | `dist` | ビルド出力先 (これを配信する) |
| `-p` / `--port` | `4321` | ポート番号 |

```bash
roadmapper dev -c my-roadmap/roadmap.yml --port 8080
```

## 学ぶこと

- **validate --strict** — warning も失敗として扱い、CI で記事の欠落を検出する方法

## サブタスク

- [ ] `validate` でエラーゼロを確認した
- [ ] `dev` コマンドでブラウザに表示できた
- [ ] `roadmap.yml` を編集してライブリロードを体験した
