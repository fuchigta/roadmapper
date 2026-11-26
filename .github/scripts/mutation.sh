#!/usr/bin/env bash
# ミューテーションテスト (gremlins) を基準 ref との差分に対して実行し、生き残り数でゲートする。
# CI (.github/workflows/mutation.yml) と手元で同じ手順を使う。
#
#   .github/scripts/mutation.sh origin/master   # 差分実行 (CI と同じ)
#   .github/scripts/mutation.sh --all           # 対象パッケージ全体 (ベースライン計測用。CI では使わない)
#
# gremlins v0.6.0 を PATH に置いておくこと:
#   go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
#
# ---- 設計メモ (gremlins v0.6.0 の実挙動で確認した注意点) ----
# - 差分実行は gremlins の --diff <ref> を使う。中身は `git diff --merge-base <ref>` で、
#   追加・変更行に乗ったミュータントだけを実行し、それ以外は SKIPPED (集計対象外) になる
# - --diff は「差分が空なら全行を変更扱い」にする。全件実行にならないよう、対象パッケージの
#   非テスト .go に差分が無ければ gremlins を呼ばずに終了する
# - --diff のファイル名照合は「gremlins に渡したパスからの相対パス」と git diff の
#   リポジトリルート相対パスの比較なので、パッケージを直接渡すと一致せず全件 SKIPPED になる。
#   そのためモジュールルート (.) を渡し、対象外のディレクトリを --exclude-files (ファイル単位の
#   正規表現。除外したファイルは集計からも外れる) で落とす
# - gremlins は hunk ごとに「先頭の追加行から追加行数ぶん」を変更範囲とみなすため、既定の
#   前後 3 行コンテキストだと 1 つの hunk に複数の変更塊があるとき 2 つ目以降を取りこぼす。
#   diff.context=0 を環境変数経由の git 設定で渡し、変更塊ごとに hunk を分ける
# - 各ミュータントのタイムアウトは「カバレッジ計測 (go test -cover ./...) の所要時間 × 係数」。
#   テストキャッシュが効くと所要時間が 1 秒程度まで縮み、既定係数 3 では正常なテストまで
#   TIMED OUT になる (TIMED OUT は KILLED/LIVED のどちらにも数えられず、生き残りを隠す)。
#   GOFLAGS=-count=1 でキャッシュを使わせず所要時間を実テスト時間 (手元で約 5 秒) に揃え、
#   係数 10 で余裕を持たせる
set -euo pipefail

# ---- 対象範囲 ----
# 純粋関数でテストが厚いパッケージから始める。広げるときは下のベースラインを測り直してゲートも見直す
TARGETS=(internal/changelog internal/graph)

# ---- ゲート ----
# 生き残り = LIVED + NOT COVERED (どちらもテストで検出されないミュータント)。
# ベースライン (2026-11-26, gremlins v0.6.0, Go 1.25.6, --all で対象全体):
#   internal/changelog: KILLED 22 / LIVED 2 / NOT COVERED 0 (efficacy 91.67%)
#   internal/graph:     KILLED 8  / LIVED 0 / NOT COVERED 0 (efficacy 100%)
#   合計 KILLED 30 / 生き残り 2 / TIMED OUT 0 (efficacy 93.75%)、ミューテーション実行 約 25 秒・
#   カバレッジ計測込みで約 35 秒
# 生き残り 2 件はどちらも等価ミュータント: changelog.go の Check 末尾のソート比較
#   `ws[i].NodeID < ws[j].NodeID` と `ws[i].Message < ws[j].Message` の `<` → `<=`
#   (CONDITIONALS_BOUNDARY)。前者は `!=` の分岐内で等号が成り立たず、後者は NodeID・Message が
#   等しい Warning は区別できないため出力が変わらない
# Recent は PR #12 で暦日判定に変わり、旧実装の `now.Sub(d) <= limit` のミュータントは
# 消滅した (After/Before のメソッド呼び出しは gremlins の変異対象外)。その境界は
# TestRecentBoundary がテストで固定している
# gremlins にはミュータント単位・行単位の除外が無く (除外はファイル単位の --exclude-files と
# 変異種別単位のフラグのみで、どちらも検出力を落とす)、等価ミュータントは除外せず上限で吸収する。
# 差分実行でも最悪 (対象全体に差分がある) の値は実測の 2。新しい等価ミュータント 1 件ぶんの余裕を
# 足して 3 に置くラチェットとし、テストを足して実測が下がったら引き下げる
MAX_SURVIVING_MUTANTS=3
TIMEOUT_COEFFICIENT=10

usage() {
  echo "usage: $0 <base-ref> | --all" >&2
  exit 2
}

[ $# -eq 1 ] || usage
base="$1"

cd "$(git rev-parse --show-toplevel)"

diff_args=()
if [ "$base" != "--all" ]; then
  changed=$(git diff --name-only --diff-filter=AMR --merge-base "$base" -- "${TARGETS[@]}" |
    grep -E '\.go$' | grep -v -E '_test\.go$' || true)
  if [ -z "$changed" ]; then
    echo "対象パッケージ (${TARGETS[*]}) の非テスト .go に $base からの差分が無いため、ミューテーションテストを省略します"
    exit 0
  fi
  echo "差分のあるファイル:"
  printf '  %s\n' $changed
  diff_args=(--diff "$base")
fi

# 対象外のディレクトリ (非テスト .go を含むもの) をファイル単位で除外する
exclude_args=()
while IFS= read -r dir; do
  skip=0
  for t in "${TARGETS[@]}"; do
    [ "$dir" = "$t" ] && skip=1
  done
  [ "$skip" -eq 1 ] && continue
  exclude_args+=(--exclude-files "^${dir//./\\.}/[^/]+\$")
done < <(find . -name '*.go' -not -name '*_test.go' -not -path './.git/*' -printf '%h\n' | sed 's|^\./||' | sort -u)

report=$(mktemp)
trap 'rm -f "$report"' EXIT

GOFLAGS=-count=1 \
  GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=diff.context GIT_CONFIG_VALUE_0=0 \
  gremlins unleash \
  --timeout-coefficient "$TIMEOUT_COEFFICIENT" \
  --output-statuses lct \
  --output "$report" \
  "${diff_args[@]}" "${exclude_args[@]}" .

killed=$(jq '.mutants_killed' "$report")
lived=$(jq '.mutants_lived' "$report")
not_covered=$(jq '.mutants_not_covered' "$report")
timed_out=$(jq '[.files[].mutations[] | select(.status == "TIMED OUT")] | length' "$report")
surviving=$((lived + not_covered))

echo
echo "KILLED $killed / LIVED $lived / NOT COVERED $not_covered / TIMED OUT $timed_out"
echo "生き残り (LIVED + NOT COVERED): $surviving (上限 $MAX_SURVIVING_MUTANTS)"
if [ "$timed_out" -gt 0 ]; then
  echo "注意: TIMED OUT は KILLED 扱い。無限ループ以外で出ているならタイムアウトが短すぎる (TIMEOUT_COEFFICIENT を見直す)"
fi
if [ "$surviving" -gt "$MAX_SURVIVING_MUTANTS" ]; then
  echo "失敗: 生き残りミュータントが上限を超えています。上の LIVED / NOT COVERED を殺すテストを追加してください" >&2
  exit 1
fi
