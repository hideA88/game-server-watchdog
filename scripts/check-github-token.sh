#!/bin/bash

echo "=== GitHub Token チェックツール ==="
echo

# 1. 環境変数の確認
if [ -z "$GITHUB_TOKEN" ]; then
    echo "❌ GITHUB_TOKEN が設定されていません"
    echo ""
    echo "設定方法:"
    echo "  export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    echo ""
    echo "または ~/.bashrc や ~/.zshrc に追加してください"
    exit 1
else
    echo "✅ GITHUB_TOKEN が設定されています"
    # トークンの最初の文字だけ表示（セキュリティのため）
    echo "   トークン: ghp_${GITHUB_TOKEN:4:4}...(残りは非表示)"
fi

echo
echo "=== トークンの権限チェック ==="

# 2. GitHub API でトークンを検証
echo -n "ユーザー情報の取得... "
USER_INFO=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user)

if echo "$USER_INFO" | grep -q '"login"'; then
    USERNAME=$(echo "$USER_INFO" | grep '"login"' | cut -d'"' -f4)
    echo "✅ 成功 (ユーザー: $USERNAME)"
else
    echo "❌ 失敗"
    echo "エラー: $USER_INFO"
    exit 1
fi

# 3. トークンタイプの判定とスコープ確認
echo
echo "=== トークンタイプとスコープ確認 ==="

# レスポンスヘッダーを取得
HEADERS=$(curl -sI -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user)
SCOPES=$(echo "$HEADERS" | grep -i "x-oauth-scopes:" | cut -d' ' -f2-)
TOKEN_TYPE=$(echo "$HEADERS" | grep -i "x-github-authentication-token-expiration:" | cut -d' ' -f2-)

if [ -n "$TOKEN_TYPE" ]; then
    echo "✅ Fine-grained personal access token を使用中"
    echo "   有効期限: $TOKEN_TYPE"
    echo ""
    echo "Fine-grained トークンはスコープではなく権限ベースです。"
    echo "必要な権限:"
    echo "  - Contents: Read and write"
    echo "  - Actions: Read"
    echo "  - Metadata: Read"
elif [ -z "$SCOPES" ]; then
    echo "⚠️  トークンタイプを判定できませんでした"
    echo "   Fine-grained personal access token の可能性があります"
else
    echo "Classic personal access token を使用中"
    echo "現在のスコープ: $SCOPES"
    echo
    echo "必要なスコープのチェック:"
    
    # repo スコープ
    if echo "$SCOPES" | grep -q "repo"; then
        echo "✅ repo - リポジトリへのフルアクセス"
    else
        echo "❌ repo - リリース作成に必要です"
    fi
fi

# 4. リポジトリへのアクセステスト
echo
echo "=== リポジトリアクセステスト ==="
echo -n "hideA88/game-server-watchdog リポジトリへのアクセス... "

REPO_INFO=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/repos/hideA88/game-server-watchdog)

if echo "$REPO_INFO" | grep -q '"full_name"'; then
    echo "✅ 成功"
    
    # push権限の確認
    if echo "$REPO_INFO" | grep -q '"push": true'; then
        echo "✅ プッシュ権限: あり"
    else
        echo "❌ プッシュ権限: なし（リリース作成に必要）"
    fi
else
    echo "❌ 失敗"
    echo "エラー: リポジトリにアクセスできません"
fi

# 5. リリース作成のテスト（ドライラン）
echo
echo "=== リリース作成権限のテスト ==="
echo -n "リリース一覧の取得... "

RELEASES=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/repos/hideA88/game-server-watchdog/releases)

if echo "$RELEASES" | grep -q '\[' || echo "$RELEASES" | grep -q '"message": "Not Found"'; then
    echo "✅ 成功（リリースAPIにアクセス可能）"
else
    echo "❌ 失敗"
    echo "エラー: $RELEASES"
fi

# 6. レート制限の確認
echo
echo "=== API レート制限 ==="
RATE_LIMIT=$(curl -s -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/rate_limit)
REMAINING=$(echo "$RATE_LIMIT" | grep -A2 '"core"' | grep '"remaining"' | grep -o '[0-9]*')
LIMIT=$(echo "$RATE_LIMIT" | grep -A2 '"core"' | grep '"limit"' | grep -o '[0-9]*' | head -1)

if [ -n "$REMAINING" ] && [ -n "$LIMIT" ]; then
    echo "API残り回数: $REMAINING / $LIMIT"
else
    echo "⚠️  レート制限情報を取得できませんでした"
fi

# 7. 実際のリリース作成テスト
echo
echo "=== リリース作成テスト（ドライラン）==="
echo -n "リリース作成権限の詳細チェック... "

# テスト用のリリースデータ
TEST_RELEASE_DATA='{
  "tag_name": "v0.0.0-test",
  "name": "Test Release (dry-run)",
  "body": "This is a test release",
  "draft": true,
  "prerelease": true
}'

# ドライランとして draft リリースを作成を試みる
CREATE_RESPONSE=$(curl -s -X POST \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  -d "$TEST_RELEASE_DATA" \
  https://api.github.com/repos/hideA88/game-server-watchdog/releases)

if echo "$CREATE_RESPONSE" | grep -q '"draft": true'; then
    echo "✅ 成功"
    echo "   リリース作成権限が確認できました"
    
    # 作成したドラフトを削除
    RELEASE_ID=$(echo "$CREATE_RESPONSE" | grep '"id"' | head -1 | grep -o '[0-9]*')
    if [ -n "$RELEASE_ID" ]; then
        curl -s -X DELETE \
          -H "Authorization: token $GITHUB_TOKEN" \
          https://api.github.com/repos/hideA88/game-server-watchdog/releases/$RELEASE_ID > /dev/null
    fi
elif echo "$CREATE_RESPONSE" | grep -q "Not Found"; then
    echo "❌ 失敗"
    echo "   エラー: リポジトリが見つかりません"
elif echo "$CREATE_RESPONSE" | grep -q "requires authentication"; then
    echo "❌ 失敗"
    echo "   エラー: 認証に失敗しました"
elif echo "$CREATE_RESPONSE" | grep -q "Resource not accessible"; then
    echo "❌ 失敗"
    echo "   エラー: リソースにアクセスできません（権限不足）"
else
    echo "❌ 失敗"
    echo "   エラー: $CREATE_RESPONSE"
fi

# 8. 推奨事項
echo
echo "=== 推奨事項 ==="

# Fine-grained token の場合
if [ -n "$TOKEN_TYPE" ]; then
    echo "✅ Fine-grained personal access token を使用しています"
    echo ""
    if echo "$CREATE_RESPONSE" | grep -q '"draft": true'; then
        echo "✅ GitHub Releases の作成に必要な権限があります！"
        echo ""
        echo "次のステップ:"
        echo "1. make release-dry-run  # ローカルでテスト"
        echo "2. make tag VERSION=v0.1.0  # タグ作成"
        echo "3. git push origin v0.1.0  # リリース開始"
    else
        echo "⚠️  リリース作成権限が不足している可能性があります"
        echo "   以下の権限が設定されているか確認してください:"
        echo "   - Contents: Read and write"
        echo "   - Actions: Read"
        echo "   - Metadata: Read"
    fi
# Classic token の場合
elif [ -n "$SCOPES" ]; then
    if echo "$SCOPES" | grep -q "repo"; then
        echo "✅ GitHub Releases の作成に必要な権限があります！"
        echo ""
        echo "次のステップ:"
        echo "1. make release-dry-run  # ローカルでテスト"
        echo "2. make tag VERSION=v0.1.0  # タグ作成"
        echo "3. git push origin v0.1.0  # リリース開始"
    else
        echo "⚠️  'repo' スコープが不足しています。GitHub Releases を作成するには必要です。"
        echo "   新しいトークンを作成してください: https://github.com/settings/tokens/new"
    fi
else
    echo "⚠️  トークンタイプを判定できませんでした"
    echo "   トークンが正しく設定されているか確認してください"
fi