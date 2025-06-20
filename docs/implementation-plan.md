# Game Server Watchdog 実装計画

## 概要
Docker Compose 環境でゲームサーバーを統合監視するシステムの構築

## 新機能一覧

### 1. 拡張監視コマンド

#### `@bot monitor` - リアルタイム監視ダッシュボード
```
🖥️ **システム監視ダッシュボード**

📊 **ホストサーバー**
CPU: ████████░░ 80% (32 cores)
MEM: ██████░░░░ 60% (64GB/128GB)
DISK: ███░░░░░░░ 30% (300GB/1TB)

📦 **コンテナ状況**
┌─────────────────┬────────┬────────┬────────┬────────┐
│ サービス         │ 状態   │ CPU    │ メモリ │ 稼働   │
├─────────────────┼────────┼────────┼────────┼────────┤
│ minecraft       │ 🟢     │ 25%    │ 8GB    │ 2d 5h  │
│ rust           │ 🟢     │ 45%    │ 32GB   │ 1d 12h │
│ ark-island     │ 🔴     │ -      │ -      │ -      │
│ ark-scorched   │ 🟡     │ 90%    │ 30GB   │ 5h 20m │
└─────────────────┴────────┴────────┴────────┴────────┘

⚠️ **アラート**
- ark-scorched: CPU使用率が高い (90%)
```

#### `@bot container <name>` - 個別コンテナ詳細
```
@bot container rust
```

#### `@bot restart <name>` - コンテナ再起動
```
@bot restart minecraft
```

#### `@bot logs <name> [lines]` - ログ表示
```
@bot logs rust 50
```

### 2. 自動化機能

#### ヘルスチェック
- 5分ごとにコンテナの状態を確認
- 異常検知時にDiscordに通知

#### リソース管理
- CPU/メモリ閾値超過時の自動アクション
- 設定可能な閾値（環境変数）

#### スケジュール機能
```yaml
# 環境変数での設定例
WATCHDOG_SCHEDULE_ENABLED=true
WATCHDOG_SCHEDULE_RESTART_TIMES=03:00,15:00  # 毎日3時と15時に再起動
```

### 3. Docker Compose 統合

#### 使用方法
```bash
# 単体起動
docker compose -f docker-compose.yml up -d

# Watchdog 込みで起動
docker compose -f docker-compose.yml -f docker-compose.watchdog.yml up -d
```

#### 環境変数
```env
# Watchdog 専用設定
WATCHDOG_CHECK_INTERVAL=300  # ヘルスチェック間隔（秒）
WATCHDOG_CPU_THRESHOLD=85    # CPU閾値（%）
WATCHDOG_MEM_THRESHOLD=90    # メモリ閾値（%）
WATCHDOG_AUTO_RESTART=true   # 自動再起動有効化
WATCHDOG_ALERT_CHANNEL=${ALLOWED_CHANNEL_IDS}  # アラート送信先
```

## 実装手順

### Phase 1: 基本機能（1週間）
- [x] Docker Compose 設定ファイル作成
- [ ] Dockerfile の最適化
- [ ] ホストシステム情報取得機能
- [ ] コンテナリソース監視機能

### Phase 2: 監視機能（1週間）
- [ ] `monitor` コマンド実装
- [ ] `container` コマンド実装
- [ ] リアルタイムメトリクス収集
- [ ] Discord での表示最適化

### Phase 3: 管理機能（1週間）
- [ ] `restart` コマンド実装
- [ ] `logs` コマンド実装
- [ ] 自動再起動機能
- [ ] スケジュール機能

### Phase 4: 高度な機能（2週間）
- [ ] Prometheus メトリクスエクスポート
- [ ] Grafana ダッシュボード連携
- [ ] Web UI（オプション）
- [ ] バックアップ/リストア機能

## アーキテクチャ

```
┌─────────────────────────────────────────────────────┐
│                   Host Server                        │
│                                                      │
│  ┌─────────────────────────────────────────────┐   │
│  │            Docker Compose Network             │   │
│  │                                               │   │
│  │  ┌───────────┐ ┌───────────┐ ┌───────────┐  │   │
│  │  │Minecraft  │ │   Rust    │ │    ARK    │  │   │
│  │  │Container  │ │Container  │ │Containers │  │   │
│  │  └───────────┘ └───────────┘ └───────────┘  │   │
│  │         ↑             ↑             ↑         │   │
│  │         └─────────────┴─────────────┘         │   │
│  │                       │                       │   │
│  │               ┌───────────────┐               │   │
│  │               │   Watchdog    │               │   │
│  │               │   Container   │               │   │
│  │               └───────┬───────┘               │   │
│  │                       │                       │   │
│  └───────────────────────┼───────────────────────┘   │
│                         │                           │
│                    Docker Socket                    │
│                    /var/run/docker.sock            │
└─────────────────────────────────────────────────────┘
                          │
                          ↓
                    Discord Bot API
```

## セキュリティ考慮事項

1. **Docker ソケット**
   - 読み取り専用でマウント
   - 最小権限の原則

2. **リソース制限**
   - Watchdog 自体のリソース制限
   - DoS 攻撃対策

3. **アクセス制御**
   - Discord のチャンネル/ユーザー制限
   - コマンドごとの権限設定

## 今後の拡張性

1. **マルチホスト対応**
   - Docker Swarm 対応
   - Kubernetes 対応

2. **プラグインシステム**
   - カスタムヘルスチェック
   - ゲーム固有の監視

3. **外部連携**
   - Datadog
   - New Relic
   - CloudWatch