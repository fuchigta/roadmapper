## validate — 設定ファイルのチェック

ビルド前に `roadmap.yml` の構文・循環参照・重複 ID などを検証します。

```bash
roadmapper validate -c my-roadmap/roadmap.yml
```

エラーがなければ `OK` と表示されます。

## dev — ライブプレビュー

ファイルを保存するたびにブラウザが自動更新される開発サーバです。

```bash
roadmapper dev -c my-roadmap/roadmap.yml
# http://localhost:4321 で起動
```

`--port` オプションでポートを変更できます。

```bash
roadmapper dev -c my-roadmap/roadmap.yml --port 8080
```

## サブタスク

- [ ] `validate` でエラーゼロを確認した
- [ ] `dev` コマンドでブラウザに表示できた
- [ ] `roadmap.yml` を編集してライブリロードを体験した
