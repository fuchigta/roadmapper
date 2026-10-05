---
links:
  - { title: "systemd 公式サイト", url: "https://systemd.io/" }
  - { title: "man7.org: systemctl(1)", url: "https://man7.org/linux/man-pages/man1/systemctl.1.html" }
---

## 学ぶこと

systemd は多くの Linux ディストリビューションで採用されているサービス管理の仕組みです。
サービスの起動・停止・自動起動や、ログの確認ができるようになりましょう。

## サブタスク

- [ ] `systemctl` でサービスの起動・停止・状態確認ができる
- [ ] サービスの自動起動を有効・無効にできる
- [ ] `journalctl` でサービスのログを確認できる
- [ ] 簡単なユニットファイルを書ける

## ポイント

```bash
systemctl status nginx
sudo systemctl enable --now nginx
journalctl -u nginx --since today
```
