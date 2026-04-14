---
links:
  - { title: "roadmap.yml スキーマ説明 (README)", url: "https://github.com/fuchigta/roadmapper#roadmapyml" }
---

## init コマンドでスケルトン生成

`roadmapper init` コマンドは、すぐに動くサンプルプロジェクトを生成します。

```bash
# minimal テンプレート (最小構成)
roadmapper init my-roadmap --template minimal

# blank テンプレート (roadmap.yml のスケルトンのみ・最速スタート)
roadmapper init my-roadmap --template blank

# frontend-beginner テンプレート (フロントエンド学習ロードマップ)
roadmapper init my-roadmap --template frontend-beginner

# backend-beginner テンプレート (バックエンド学習ロードマップ)
roadmapper init my-roadmap --template backend-beginner

# devops テンプレート (DevOps/インフラロードマップ)
roadmapper init my-roadmap --template devops
```

| テンプレート | 用途 |
|---|---|
| `minimal` | 最小構成のサンプル。構造を理解するのに最適 |
| `blank` | `roadmap.yml` のスケルトンのみ。ゼロから書き始めたい場合 |
| `frontend-beginner` | フロントエンド開発の学習ロードマップ (複数ロードマップの例を含む) |
| `backend-beginner` | バックエンド開発の学習ロードマップ |
| `devops` | DevOps/インフラエンジニアリングのロードマップ |

生成されるディレクトリ構造:

```
my-roadmap/
├── roadmap.yml        # ロードマップ定義
└── content/           # 各ノードの詳細説明 (Markdown)
    └── html.md
```

## サブタスク

- [ ] `roadmapper init` でプロジェクトを作成した
- [ ] `roadmap.yml` の中身を確認した
- [ ] `content/` ディレクトリを確認した
