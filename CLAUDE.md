# Game Server Watchdog

このプロジェクトはゲームサーバーを監視・管理するDiscord botです。

## プロジェクト概要

- **言語**: Go
- **用途**: ゲームサーバーの監視、Discord経由での管理
- **主要ライブラリ**: discordgo

## 重要な設定

### 環境変数 (.env)
- `DISCORD_TOKEN`: Discord botトークン（必須）
- `ALLOWED_CHANNEL_IDS`: 許可されたチャンネルID（カンマ区切り）
- `ALLOWED_USER_IDS`: 許可されたユーザーID（カンマ区切り）

### アクセス制限
- 特定のチャンネルIDのみでボットが反応するよう制限済み
- メンション必須：すべてのコマンドはボットをメンションして実行

## 実装済みコマンド

- `@bot ping`: 応答確認（pong!!を返す）
- `@bot help`: ヘルプメッセージを表示

## 開発時の注意事項

### コマンド実行
- `make lint`: コードのlintチェック
- `go run cmd/watchdog/main.go`: ボット起動

### 新しいハンドラー追加時
1. `internal/bot/handler/`に新しいハンドラーファイルを作成
2. `handler.NewHandlers()`に追加
3. メンションチェックを必ず実装

### セキュリティ
- Discord tokenは絶対にコミットしない
- `.env`ファイルは`.gitignore`に含める

## 今後の実装予定
- サーバー監視機能（CPU、メモリ、ディスク使用率）
- ゲームサーバーの起動/停止制御
- Minecraft RCON対応（既に基盤あり）
- アラート通知機能