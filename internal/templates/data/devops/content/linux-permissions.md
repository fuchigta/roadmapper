---
links:
  - { title: "man7.org: chmod(1)", url: "https://man7.org/linux/man-pages/man1/chmod.1.html" }
  - { title: "man7.org: path_resolution(7)", url: "https://man7.org/linux/man-pages/man7/path_resolution.7.html" }
---

## 学ぶこと

Linux ではすべてのファイルに所有者・グループ・パーミッションが設定されています。
権限を正しく理解することは、安全なサーバー運用の基本です。

## サブタスク

- [ ] `ls -l` の出力から権限を読み取れる
- [ ] `chmod` / `chown` で権限と所有者を変更できる
- [ ] `/etc` `/var` `/home` など主要ディレクトリの役割を説明できる
- [ ] `sudo` を最小限に使う理由を理解している

## ポイント

```bash
ls -l script.sh
chmod 750 script.sh   # 所有者: rwx, グループ: r-x, その他: なし
```
