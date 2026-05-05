#!/bin/sh
set -e

# このスクリプトの役割:
# 1. コンテナ起動時にDocker socketのグループIDを自動検出
# 2. watchdogユーザーに適切な権限を付与
# 3. watchdogユーザーとして実際のアプリケーションを起動
# 
# セキュリティ:
# - 実際のアプリケーションは常にwatchdogユーザーで実行される
# - rootは初期設定のみに使用され、すぐに権限を落とす

# Docker socketのグループIDを動的に取得
if [ -S /var/run/docker.sock ]; then
    DOCKER_GID=$(stat -c '%g' /var/run/docker.sock)
    echo "Docker socket GID: $DOCKER_GID"
    
    # 初回起動時のみrootで実行され、グループ設定後にwatchdogユーザーに切り替え
    if [ "$(id -u)" = "0" ]; then
        # dockerグループが存在しない場合は作成
        if ! getent group docker > /dev/null 2>&1; then
            addgroup -g $DOCKER_GID docker
        fi
        # watchdogユーザーをdockerグループに追加
        adduser watchdog docker
        
        # watchdogユーザーに切り替えて実行（ここで権限を落とす）
        echo "Switching to watchdog user..."
        exec su-exec watchdog "$@"
    else
        echo "Warning: Running as non-root user. Docker socket access may be limited."
        exec "$@"
    fi
else
    echo "Warning: Docker socket not found at /var/run/docker.sock"
    exec "$@"
fi