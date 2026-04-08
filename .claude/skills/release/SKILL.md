---
name: release
description: >
  roadmapper をリリースするとき、タグを切るとき、「リリースして」「バージョンを上げて」と
  言われたときに使う。バージョン決定 → タグ作成 → CI 監視 → 失敗時リカバリ まで一貫して行う。
---

# roadmapper リリース手順

必ずこのスキルに定めた手順を守る。省略・変更は禁止。

---

## Phase 0: 事前チェック

以下を **順番に** 実行する。問題があれば次フェーズに進む前に解決する。

### 0-1. gh CLI 認証確認
```bash
gh auth status
```
未認証なら中断してユーザーに `gh auth login` を促す。

### 0-2. ブランチ確認
```bash
git branch --show-current
```
`master` 以外なら **中断してエラー報告**。スキルを終了する。

### 0-3. 作業ツリーの確認と自動コミット
```bash
git status --porcelain
```
変更がある場合:
1. `git diff HEAD` で変更内容を確認する
2. 変更内容からコンベンショナルコミットの type と description を推測する
3. ユーザー確認なしでコミットする:
   ```bash
   git add -A
   git commit -m "<推測した type>(scope): <description>"
   ```
4. lefthook の `commit-msg` フックが形式を検証する。通過しなければ修正して再試行。

### 0-4. リモート同期
```bash
git fetch origin master
```
- **ローカルが進んでいる (未プッシュコミットあり)**: `git push origin master`
- **origin が進んでいる (他者プッシュ)**: `git pull --ff-only origin master` を試す
  - fast-forward 不可なら **中断してユーザーにエスカレート** (コンフリクト解消を依頼)
- **両方進んでいる**: 中断してユーザーにエスカレート

---

## Phase 1: バージョン決定

### 1-1. 前回タグと差分コミットを取得
```bash
git describe --tags --abbrev=0
git log <prev-tag>..HEAD --pretty=format:"%s%n%b" --no-merges
```

### 1-2. バージョン算出ルール

現在は **0.x 系**。以下のルールで bump 種別を決める:

| 条件 | bump |
|---|---|
| コミット本文に `BREAKING CHANGE:` / type に `!` が含まれる | **minor** (0.x → 0.(x+1).0) |
| `feat:` または `feat(...):` を含む | **minor** |
| `fix:` のみ | **patch** (0.x.y → 0.x.(y+1)) |
| `chore:` / `docs:` / `ci:` / `style:` / `refactor:` / `perf:` / `test:` のみ | **patch** |

> **1.0.0 以降に達したら**: `BREAKING CHANGE` → major、`feat` → minor、その他 → patch に切り替える。

### 1-3. ユーザー承認

算出したバージョンと根拠コミット一覧を表示し、AskUserQuestion で確認:
- 「このバージョンで進める」
- 「別のバージョンを指定する」(ユーザーが手入力)

承認を得てから次フェーズへ進む。

---

## Phase 2: タグ作成 & プッシュ

```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```

push 失敗時はローカルタグを削除して中断:
```bash
git tag -d vX.Y.Z
```

---

## Phase 3: CI ウォッチ

タグ push 直後は run がまだ存在しない場合がある。以下の手順で ID を取得:

1. `release.yml` の run ID を取得 (pending になるまで最大 30 秒リトライ):
   ```bash
   gh run list --workflow=release.yml --limit 1 --json databaseId,status,headSha
   ```
2. `pages.yml` の run ID を取得 (pages は master push でトリガー。0-4 の master push 後に発生):
   ```bash
   gh run list --workflow=pages.yml --limit 1 --json databaseId,status,headSha
   ```
3. **両方を並列で監視**:
   ```bash
   gh run watch <release-run-id> --exit-status
   gh run watch <pages-run-id> --exit-status
   ```
4. 両方成功 → Phase 5 へ
5. いずれかが失敗 → Phase 4 へ

---

## Phase 4: 失敗時リカバリ

**最大 2 回まで** リトライする (`attempt` カウンタで管理)。上限に達したら中断してユーザーにエスカレート。

### ループ処理 (attempt = 0, 1)

1. 失敗ログを取得:
   ```bash
   gh run view <failed-run-id> --log-failed
   ```
2. 失敗原因を要約してユーザーに表示する
3. 修正方針を判断:
   - GoReleaser ビルドエラー → `.goreleaser.yml` や Go コードを修正
   - テスト失敗 → 該当テストを修正
   - Pages ビルドエラー → `docs/roadmap.yml` 等を確認
4. 修正をコミット (Phase 0 の 0-3 と同じ規則: 変更から推測した conventional commit)
5. master に push:
   ```bash
   git push origin master
   ```
6. **次のバージョン番号で** タグを切って push (Phase 1 のルールで 1 段 bump):
   - 失敗したタグはリモート/ローカルとも **そのまま残す** (削除禁止)
   - 例: v0.3.0 が失敗 → fix コミット追加 → v0.3.1 でリリース
7. Phase 3 に戻る

### 上限超過時
`attempt >= 2` で再度失敗したら:
- これまでの失敗したタグ一覧と失敗ログを提示
- 「自動リカバリの上限 (2 回) に達しました。手動での対応が必要です。」と報告
- スキルを終了する

---

## Phase 5: 成功報告

以下を表示してスキルを終了する:

- GitHub Release URL: `https://github.com/fuchigta/roadmapper/releases/tag/vX.Y.Z`
- GitHub Pages URL (pages.yml の output から取得できれば)
- このリリースに含まれる変更の要約 (Phase 1 で取得したコミット一覧から)

---

## 禁止事項

- `git tag -f` (タグの強制上書き) 禁止
- `git push --force` 禁止
- 失敗したタグのリモート削除禁止 (`git push origin :vX.Y.Z` 等)
- Phase 0-2 のブランチ確認をスキップする禁止
- ユーザー承認 (Phase 1-3) をスキップして自動でタグを作成することは禁止
