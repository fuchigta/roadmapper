#!/bin/bash
# wait-for-run.sh <workflow> [sha] [timeout_seconds]
#
# 指定ワークフローの run が現れるまで待ち、databaseId を標準出力に返す。
# タイムアウト時は exit 1。
#
# 引数:
#   workflow        — ワークフローファイル名 (例: release.yml)
#   sha             — 対象コミット SHA (省略時は HEAD)
#   timeout_seconds — タイムアウト秒数 (省略時: 30)

WORKFLOW="${1:?usage: wait-for-run.sh <workflow> [sha] [timeout_seconds]}"
SHA="${2:-$(git rev-parse HEAD)}"
TIMEOUT="${3:-30}"

INTERVAL=5
elapsed=0

while [ "$elapsed" -lt "$TIMEOUT" ]; do
  RUN_ID=$(gh run list --workflow="$WORKFLOW" --limit 5 \
    --json databaseId,headSha \
    --jq ".[] | select(.headSha == \"$SHA\") | .databaseId" 2>/dev/null | head -1)

  if [ -n "$RUN_ID" ]; then
    echo "$RUN_ID"
    exit 0
  fi

  sleep "$INTERVAL"
  elapsed=$((elapsed + INTERVAL))
done

echo "ERROR: workflow '$WORKFLOW' の run が ${TIMEOUT}s 以内に見つかりませんでした (sha=$SHA)" >&2
exit 1
