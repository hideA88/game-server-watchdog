#!/bin/bash

VERSION=${1:-"latest"}

echo "=== リリース確認スクリプト ==="
echo "バージョン: $VERSION"
echo

# GitHub Release の確認
echo "1. GitHub Release の確認..."
if command -v gh &> /dev/null; then
    gh release view $VERSION --repo hideA88/game-server-watchdog 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "✅ GitHub Release が作成されています"
    else
        echo "❌ GitHub Release が見つかりません"
    fi
else
    echo "⚠️  GitHub CLI がインストールされていません"
    echo "   https://github.com/hideA88/game-server-watchdog/releases で確認してください"
fi

echo

# Docker イメージの確認
echo "2. Docker イメージの確認..."
if docker pull ghcr.io/hidea88/game-server-watchdog:$VERSION 2>&1 | grep -q "Downloaded newer image\|Image is up to date"; then
    echo "✅ Docker イメージが利用可能です"
    docker images | grep game-server-watchdog
else
    echo "❌ Docker イメージのプルに失敗しました"
fi

echo

# バイナリのダウンロードテスト
echo "3. バイナリのダウンロードテスト..."
DOWNLOAD_URL="https://github.com/hideA88/game-server-watchdog/releases/download/$VERSION/game-server-watchdog_${VERSION#v}_linux_x86_64.tar.gz"
if curl -sL --fail -o /dev/null "$DOWNLOAD_URL"; then
    echo "✅ Linux x86_64 バイナリがダウンロード可能です"
    echo "   URL: $DOWNLOAD_URL"
else
    echo "❌ バイナリのダウンロードに失敗しました"
fi

echo
echo "=== インストールコマンド例 ==="
echo
echo "# Docker を使う場合:"
echo "docker run --rm -it \\"
echo "  --env-file .env \\"
echo "  -v /var/run/docker.sock:/var/run/docker.sock \\"
echo "  ghcr.io/hidea88/game-server-watchdog:$VERSION"
echo
echo "# バイナリを使う場合 (Linux x86_64):"
echo "curl -sL $DOWNLOAD_URL | tar xz"
echo "./game-server-watchdog --version"