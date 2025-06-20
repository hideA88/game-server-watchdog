#!/bin/bash

echo "=== GitHub Release チェックリスト ==="
echo

# GitHub CLI がインストールされているか確認
if command -v gh &> /dev/null; then
    echo "✅ GitHub CLI がインストールされています"
    
    # 認証状態を確認
    if gh auth status &> /dev/null; then
        echo "✅ GitHub CLI が認証済みです"
    else
        echo "❌ GitHub CLI の認証が必要です"
        echo "   実行: gh auth login"
    fi
else
    echo "❌ GitHub CLI がインストールされていません"
    echo "   インストール: https://cli.github.com/"
fi

echo

# GoReleaser がインストールされているか確認
if command -v goreleaser &> /dev/null; then
    echo "✅ GoReleaser がインストールされています"
    goreleaser --version
else
    echo "❌ GoReleaser がインストールされていません"
    echo "   インストール: brew install goreleaser/tap/goreleaser"
fi

echo

# Docker が実行されているか確認
if docker info &> /dev/null; then
    echo "✅ Docker が実行中です"
    
    # ghcr.io にログインしているか確認
    if docker pull ghcr.io/hidea88/test 2>&1 | grep -q "unauthorized"; then
        echo "⚠️  ghcr.io へのログインが必要かもしれません"
        echo "   実行: echo \$GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
    else
        echo "✅ ghcr.io へのアクセスが可能です"
    fi
else
    echo "❌ Docker が実行されていません"
    echo "   Docker Desktop を起動してください"
fi

echo

# 環境変数の確認
if [ -n "$GITHUB_TOKEN" ]; then
    echo "✅ GITHUB_TOKEN が設定されています"
else
    echo "❌ GITHUB_TOKEN が設定されていません"
    echo "   実行: export GITHUB_TOKEN=your_token_here"
fi

echo
echo "=== リポジトリ設定の確認 ==="
echo "以下の設定を GitHub で確認してください："
echo
echo "1. Settings → Actions → General → Workflow permissions"
echo "   ✓ 'Read and write permissions' が選択されている"
echo
echo "2. 初回リリース後、パッケージ設定で："
echo "   https://github.com/hideA88?tab=packages"
echo "   ✓ game-server-watchdog パッケージにリポジトリアクセス権を付与"
echo
echo "3. Personal Access Token のスコープ："
echo "   ✓ write:packages"
echo "   ✓ read:packages"
echo "   ✓ repo (GitHub Actions 用)"

chmod +x scripts/check-release.sh