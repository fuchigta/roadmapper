---
links:
  - { title: "roadmapper テンプレート一覧 (ソース)", url: "https://github.com/fuchigta/roadmapper/tree/master/internal/templates/data" }
---

## テンプレートを選ぶ

`roadmapper init --template <名前>` で、用途に合わせたテンプレートを選べます。
指定を省略すると `minimal` が使われ、存在しない名前を指定するとエラーになります。

```bash
roadmapper init my-roadmap --template minimal
roadmapper init my-roadmap --template blank
roadmapper init my-roadmap --template frontend-beginner
roadmapper init my-roadmap --template backend-beginner
roadmapper init my-roadmap --template devops
```

| テンプレート | 内容 |
|---|---|
| `minimal` | 最小構成のサンプル。構造を理解するのに最適 |
| `blank` | `roadmap.yml` と `content/start.md` のみ。ゼロから書き始めたい場合 |
| `frontend-beginner` | フロントエンド学習ロードマップ。2 つのロードマップ (frontend / tools)、`links`・サブタスク・改版履歴 (changelog) の例入り |
| `backend-beginner` | バックエンド学習ロードマップ |
| `devops` | DevOps / インフラ学習ロードマップ |

## 選び方の目安

- **仕組みを知りたい** — `minimal` を生成して `roadmap.yml` と `content/` の対応を眺める
- **自分のテーマで作りたい** — `blank` から始めてノードを足していく
- **たたき台が欲しい** — 分野に近い学習用テンプレートを生成し、内容を書き換える

どのテンプレートも `roadmapper validate --strict` を通る状態で配布されています。

## サブタスク

- [ ] 用途に合うテンプレートを選んだ
- [ ] `init --template` で生成して `validate` が通ることを確認した
- [ ] 不要なノードや記事を削除・書き換えた
