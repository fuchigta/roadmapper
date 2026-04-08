---
links:
  - { title: "roadmap.yml スキーマ説明 (README)", url: "https://github.com/fuchigta/roadmapper#roadmapyml" }
---

## init コマンドでスケルトン生成

`roadmapper init` コマンドは、すぐに動くサンプルプロジェクトを生成します。

```bash
# minimal テンプレート (最小構成)
roadmapper init my-roadmap --template minimal

# frontend-beginner テンプレート (複数ロードマップのサンプル)
roadmapper init my-roadmap --template frontend-beginner
```

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
