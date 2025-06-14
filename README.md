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

3. ボットの実行:
```bash
make run
```

## 開発

このプロジェクトはGo言語で書かれており、以下のパッケージを使用しています:
- discordgo: Discord APIとの通信
- godotenv: 環境変数の管理
- システムメトリクス収集用のパッケージ（後ほど追加予定）
