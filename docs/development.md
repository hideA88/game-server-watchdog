# 開発環境セットアップ

## クイックスタート

### 方法1: Dockerコンテナで開発（推奨）

```bash
# 1. 環境変数を設定
cp infra/dev/.env.dev infra/dev/.env
# infra/dev/.envを編集してDISCORD_TOKENを設定

# 2. 開発環境を起動（ゲームサーバー + Watchdog）
make dev-up

# 3. ログを確認
make dev-logs

# 4. コード変更後の再ビルド
make dev-rebuild
```

### 方法2: ローカルで直接実行

```bash
# 1. ゲームサーバーのみ起動
docker compose -f docker-compose.dev.yml up -d minecraft ark-island

# 2. 環境変数を設定
cp .env.dev .env
# .envを編集してDISCORD_TOKENを設定

# 3. Watchdogをローカルで起動
go run cmd/watchdog/main.go
```

## 開発用構成の特徴

### docker-compose.dev.yml

- **最小構成**: Minecraft BedrockとARK Islandのみ
- **リソース削減**: 
  - Minecraft: 4GB RAM（本番: 16GB）
  - ARK: 8GB RAM（本番: 32GB）
- **簡素化設定**:
  - ARK: MODなし、自動更新なし
  - パスワードなし（ローカル開発用）

### 動作確認コマンド

```bash
# Discordで以下のコマンドを実行
@bot game-info    # ゲームサーバー一覧
@bot monitor      # リソース使用状況
@bot logs minecraft 20  # ログ確認
```

### トラブルシューティング

#### ポート競合

```bash
# 使用中のポートを確認
sudo lsof -i :19132
sudo lsof -i :7777

# 必要に応じてポート変更
# docker-compose.dev.yml で ports: を編集
```

#### メモリ不足

```bash
# Docker のメモリ制限を確認
docker system info | grep Memory

# 必要に応じて mem_limit を調整
```

## 本番環境との違い

| 項目 | 開発環境 | 本番環境 |
|------|----------|----------|
| 構成 | 2サーバー | 4サーバー |
| Minecraft RAM | 4GB | 16GB |
| ARK RAM | 8GB | 32GB |
| ARK MOD | なし | あり |
| 自動更新 | 無効 | 有効 |
| パスワード | なし | あり |

## 開発フロー

### 1. 機能開発
```bash
# 新機能のブランチ作成
git checkout -b feature/new-command

# 開発環境を起動
make dev-up
```

### 2. 開発サイクル
```bash
# コード変更後の動作確認
make dev-rebuild  # 再ビルドして反映

# ログ確認
make dev-logs

# 状態確認
make dev-status
```

### 3. テスト実行
```bash
# ユニットテスト
make test

# lint
make lint
```

### 4. クリーンアップ
```bash
# 開発環境の停止
make dev-down

# 完全にクリーンアップ（ボリューム含む）
make dev-clean
```

## Makeコマンド一覧

| コマンド | 説明 |
|----------|------|
| `make dev-up` | 開発環境を起動（ビルド含む） |
| `make dev-down` | 開発環境を停止 |
| `make dev-logs` | Watchdogのログを表示 |
| `make dev-rebuild` | Watchdogを再ビルド＆再起動 |
| `make dev-restart` | Watchdogを再起動（ビルドなし） |
| `make dev-status` | コンテナの状態を確認 |
| `make dev-clean` | 開発環境を完全削除 |