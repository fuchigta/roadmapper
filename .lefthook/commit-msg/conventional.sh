#!/usr/bin/env bash
# コンベンショナルコミット形式を検証する
# 参考: https://www.conventionalcommits.org/

MSG_FILE="$1"
MSG=$(cat "$MSG_FILE")

PATTERN='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?(!)?: .{1,}'

if ! echo "$MSG" | grep -qE "$PATTERN"; then
  echo ""
  echo "❌ コミットメッセージがコンベンショナルコミット形式に従っていません。"
  echo ""
  echo "  形式: <type>[(<scope>)][!]: <description>"
  echo ""
  echo "  使用可能な type:"
  echo "    feat     新機能"
  echo "    fix      バグ修正"
  echo "    docs     ドキュメントのみの変更"
  echo "    style    コードの意味に影響しない変更 (空白、フォーマット等)"
  echo "    refactor バグ修正・機能追加以外のコード変更"
  echo "    perf     パフォーマンス改善"
  echo "    test     テストの追加・修正"
  echo "    build    ビルドシステム・外部依存の変更"
  echo "    ci       CI設定の変更"
  echo "    chore    その他の変更"
  echo "    revert   コミットの取り消し"
  echo ""
  echo "  例: feat(ui): サイドパネルのリサイズ対応を追加"
  echo "      fix: ノードクリック時のクラッシュを修正"
  echo "      feat!: 設定ファイルの形式を変更 (破壊的変更)"
  echo ""
  echo "  実際のメッセージ:"
  echo "  > $MSG" | head -3
  echo ""
  exit 1
fi
