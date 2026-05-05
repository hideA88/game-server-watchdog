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
        # 同一GIDの既存グループを優先利用、無ければ docker グループを作成
        EXISTING_GROUP=$(getent group "$DOCKER_GID" | cut -d: -f1 || true)
        if [ -z "$EXISTING_GROUP" ]; then
            addgroup -g "$DOCKER_GID" docker
            EXISTING_GROUP=docker
        fi

        # watchdog ユーザーが既に該当グループに所属している場合は何もしない
        if ! id -nG watchdog | tr ' ' '\n' | grep -qx "$EXISTING_GROUP"; then
            adduser watchdog "$EXISTING_GROUP"
        fi

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
