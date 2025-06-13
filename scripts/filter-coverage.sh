#!/usr/bin/env bash
# カバレッジファイルから.coverageignoreに記載されたファイルを除外する

set -euo pipefail

COVERAGE_IN="${1:-coverage.out.tmp}"
COVERAGE_OUT="${2:-coverage.out}"
IGNORE_FILE="${3:-.coverageignore}"

# 入力ファイルが存在しない場合はエラー
if [ ! -f "$COVERAGE_IN" ]; then
    echo "Error: Coverage file $COVERAGE_IN not found" >&2
    exit 1
fi

# .coverageignoreファイルが存在しない場合は、単純にコピー
if [ ! -f "$IGNORE_FILE" ]; then
    cp "$COVERAGE_IN" "$COVERAGE_OUT"
    exit 0
fi

# grepパターンを構築
GREP_PATTERN=""
while IFS= read -r pattern || [ -n "$pattern" ]; do
    # 空行とコメント行をスキップ
    [[ -z "$pattern" || "$pattern" =~ ^[[:space:]]*# ]] && continue
    
    # パターンを追加（globパターンを正規表現に変換）
    # **/mock*.go -> .*/mock[^/]*\.go
    # *.go -> [^/]*\.go
    regex_pattern=$(echo "$pattern" | sed 's/\*\*\//.*\//g' | sed 's/\*/[^\/]*/g')
    
    if [ -n "$GREP_PATTERN" ]; then
        GREP_PATTERN="$GREP_PATTERN|$regex_pattern"
    else
        GREP_PATTERN="$regex_pattern"
    fi
done < "$IGNORE_FILE"

# フィルタリング実行
if [ -n "$GREP_PATTERN" ]; then
    # mode行は必ず残す
    head -n 1 "$COVERAGE_IN" > "$COVERAGE_OUT"
    tail -n +2 "$COVERAGE_IN" | grep -v -E "$GREP_PATTERN" >> "$COVERAGE_OUT" || true
else
    cp "$COVERAGE_IN" "$COVERAGE_OUT"
fi