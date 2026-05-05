# Game Server Watchdog

[![codecov](https://codecov.io/github/hideA88/game-server-watchdog/graph/badge.svg?token=OyRdQp2cXx)](https://codecov.io/github/hideA88/game-server-watchdog)
[![Test and Coverage](https://github.com/hideA88/game-server-watchdog/actions/workflows/test.yml/badge.svg)](https://github.com/hideA88/game-server-watchdog/actions/workflows/test.yml)

ゲームサーバーの監視と管理を行うDiscordボットです。

## 機能

- サーバーの状態監視
  - CPU使用率
  - メモリ使用率
  - ディスク使用率
  - ネットワークトラフィック
- 異常検知時の通知
  - Discordチャンネルへのアラート
  - 管理者へのメンション
- サーバー管理コマンド
  - サーバーの再起動
  - ステータス確認
  - リソース使用状況の確認

## セットアップ

1. 必要な依存関係のインストール:
```bash
make deps
```

2. 環境変数の設定:
- `.env.template`を`.env`にコピー
- Discord Bot Tokenを設定
- 監視対象のサーバー情報を設定

3. Docker権限の設定（Ubuntu/Linux環境の場合）:
```bash
# 現在のユーザーをdockerグループに追加
sudo usermod -aG docker $USER

# グループの変更を反映（再ログインが必要）
newgrp docker

# または、システムを再起動
sudo reboot
```

**注意**: dockerグループへの追加後は、セキュリティ上の理由から必ず再ログインまたは再起動を行ってください。

4. ボットの実行:
```bash
make run
```

### トラブルシューティング

#### Docker権限エラーが発生する場合
以下のようなエラーが表示される場合：
```
permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock
```

解決方法：

##### 1. ホスト上で直接実行している場合
実行ユーザーがdockerグループに属しているか確認:
```bash
groups
# dockerグループに追加
sudo usermod -aG docker $USER
newgrp docker
```

##### 2. Dockerコンテナ内から実行している場合
後述の「Docker Composeでの実行例」を参照してください。

##### 3. Docker socketの権限を確認
```bash
# ホスト上でDocker socketの権限を確認
ls -la /var/run/docker.sock
# 通常: srw-rw---- 1 root docker 0 ... /var/run/docker.sock

# dockerグループのGIDを確認
getent group docker
```

##### 4. 一時的な回避策（開発環境のみ）
```bash
# Docker socketの権限を緩める（セキュリティリスクあり）
sudo chmod 666 /var/run/docker.sock
```
**警告**: この方法は本番環境では絶対に使用しないでください。

## Docker Composeでの実行例

### セキュアな設定例（推奨）

watchdogをDocker Composeで安全に実行する場合の設定例：

#### 方法1: 自動設定（推奨）

最新版では、Docker socketのグループIDを自動的に検出して設定します：

```yaml
version: '3.8'

services:
  watchdog:
    build: .
    # または
    # image: ghcr.io/hidea88/game-server-watchdog:latest
    environment:
      - DISCORD_TOKEN=${DISCORD_TOKEN}
      - ALLOWED_CHANNEL_IDS=${ALLOWED_CHANNEL_IDS}
      - ALLOWED_USER_IDS=${ALLOWED_USER_IDS}
      - DOCKER_COMPOSE_PATH=/app/docker-compose.yml
      - DOCKER_COMPOSE_PROJECT_NAME=game-server
    volumes:
      # Docker socketをマウント（読み取り専用）
      - /var/run/docker.sock:/var/run/docker.sock:ro
      # 監視対象のdocker-compose.ymlをマウント
      - ./docker-compose.yml:/app/docker-compose.yml:ro
    # セキュリティオプション
    security_opt:
      - no-new-privileges:true
    restart: unless-stopped
```

#### 方法2: 手動でグループIDを指定

entrypointスクリプトを使用しない場合は、手動でdockerグループのGIDを指定します：

```yaml
services:
  watchdog:
    image: ghcr.io/hidea88/game-server-watchdog:latest
    environment:
      - DISCORD_TOKEN=${DISCORD_TOKEN}
      - ALLOWED_CHANNEL_IDS=${ALLOWED_CHANNEL_IDS}
      - ALLOWED_USER_IDS=${ALLOWED_USER_IDS}
      - DOCKER_COMPOSE_PATH=/app/docker-compose.yml
      - DOCKER_COMPOSE_PROJECT_NAME=game-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./docker-compose.yml:/app/docker-compose.yml:ro
    
    # watchdogユーザーで実行
    user: watchdog
    
    # ホストのdockerグループを追加（事前に`getent group docker`で確認）
    group_add:
      - 999  # dockerグループのGID
    
    # セキュリティオプション
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
    restart: unless-stopped
    
  # ゲームサーバーの例
  minecraft:
    image: itzg/minecraft-server
    labels:
      - "game.type=minecraft"
    environment:
      EULA: "TRUE"
    ports:
      - "25565:25565"
```

### セキュリティベストプラクティス

1. **非rootユーザーの使用**
   - コンテナ内でrootを使わない
   - 専用のwatchdogユーザーを作成

2. **最小権限の原則**
   - Docker socketへの読み取り専用アクセス
   - 必要最小限のファイルシステムアクセス

3. **追加のセキュリティ設定**
   - `no-new-privileges`: 権限昇格を防ぐ
   - `read_only`: ファイルシステムを読み取り専用に
   - `tmpfs`: 一時ファイル用のメモリファイルシステム

## 開発

このプロジェクトはGo言語で書かれており、以下のパッケージを使用しています:
- discordgo: Discord APIとの通信
- godotenv: 環境変数の管理
- システムメトリクス収集用のパッケージ（後ほど追加予定）
