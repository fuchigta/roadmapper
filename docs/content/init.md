---
links:
  - { title: "roadmap.yml スキーマ説明 (README)", url: "https://github.com/fuchigta/roadmapper#roadmapyml-リファレンス" }
---

## init コマンドでスケルトン生成

`roadmapper init` コマンドは、すぐに動くサンプルプロジェクトを生成します。
ディレクトリ名を省略するとカレントディレクトリに展開します。

```bash
roadmapper init my-roadmap --template frontend-beginner
```

| オプション | 説明 |
|---|---|
| `-t` / `--template` | テンプレート名 (既定: `minimal`) |

生成されるディレクトリ構造 (テンプレートによってノード数は異なります):

```
my-roadmap/
├── roadmap.yml        # ロードマップ定義
└── content/           # 各ノードの詳細説明 (Markdown)
    └── ...
```

既に存在するファイルは上書きされず、スキップされます。
生成後は `roadmapper validate` → `roadmapper build` の順に進みます。

## 学ぶこと

- **テンプレート一覧** — `--template` で選べる 5 種類のテンプレートと使い分け

## サブタスク

- [ ] `roadmapper init` でプロジェクトを作成した
- [ ] `roadmap.yml` の中身を確認した
- [ ] `content/` ディレクトリを確認した
