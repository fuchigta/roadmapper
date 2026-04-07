#!/usr/bin/env bash
set -e

unformatted=$(gofmt -l .)
if [ -n "$unformatted" ]; then
  echo "以下のファイルを gofmt してください:"
  echo "$unformatted"
  echo ""
  echo "  gofmt -w ."
  exit 1
fi
