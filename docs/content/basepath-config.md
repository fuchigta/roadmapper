## basePath とは

GitHub Pages や GitLab Pages でリポジトリ名のサブディレクトリ下に公開するとき、
CSS / JS / ページ間リンクが正しく解決されるよう `basePath` を設定します。

## 設定例

```yaml
site:
  basePath: /my-repo/    # https://user.github.io/my-repo/ で公開
```

## basePath の有無による挙動の違い

| 公開先 | basePath の設定値 |
|---|---|
| `https://user.github.io/my-repo/` | `/my-repo/` |
| `https://user.github.io/` (ユーザーサイト) | `/` または 空欄 |
| カスタムドメイン `https://example.com/` | `/` または 空欄 |

> **末尾スラッシュ必須** — `/my-repo` ではなく `/my-repo/` と書いてください。

## よくあるミス

```yaml
# NG: スラッシュが抜けている
basePath: /my-repo

# OK
basePath: /my-repo/
```

設定が間違っていると CSS が読み込まれずスタイルが崩れたり、
ページ遷移が 404 になったりします。
`validate` コマンドでは検出できないため、`build` 後にブラウザで確認してください。

## サブタスク

- [ ] `basePath` を自分のリポジトリ名に合わせて設定した
- [ ] 末尾スラッシュを含めている
- [ ] ビルド後にブラウザでリンクが正しく動作することを確認した
