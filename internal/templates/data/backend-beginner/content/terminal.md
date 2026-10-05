---
links:
  - { title: "MDN: コマンドライン入門", url: "https://developer.mozilla.org/ja/docs/Learn/Tools_and_testing/Understanding_client-side_tools/Command_line" }
---

## 学ぶこと

ターミナル (シェル) は、コマンドで OS を操作する道具です。慣れると GUI より素早く作業でき、
サーバーへ SSH で接続して作業するときにも欠かせません。

## サブタスク

- [ ] `cd` / `ls` / `pwd` でディレクトリを移動・確認できる
- [ ] `cp` / `mv` / `rm` / `mkdir` でファイルを操作できる
- [ ] `cat` / `less` / `grep` でファイルの中身を確認・検索できる
- [ ] パイプ (`|`) とリダイレクト (`>`) を使える
- [ ] `chmod` でパーミッションを変更できる

## ポイント

```bash
# ログから ERROR を含む行だけを保存する
grep ERROR app.log > errors.txt
```

> `rm` は元に戻せません。特に `-r` オプション付きで使うときは対象を確認しましょう。
