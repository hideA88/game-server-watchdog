# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1] - 2025-01-20

### Added
- 初回リリース
- Discord bot の基本機能
  - `@bot ping` - 応答確認
  - `@bot help` - ヘルプ表示
  - `@bot status` - サーバーステータス表示（CPU、メモリ、ディスク使用状況）
  - `@bot game-info` - ゲームサーバー情報表示（game.typeラベル付きコンテナのみ）
  - `@bot containers` - すべてのコンテナ一覧表示
  - `@bot restart <service>` - コンテナの再起動
  - `@bot logs <service> [lines]` - コンテナログの表示
  - `@bot monitor` - リソース使用状況のリアルタイム監視
- Docker API 連携機能
  - Docker APIを直接使用（docker-cli不要）
  - コンテナの起動/停止/再起動
  - リソース使用状況の取得
  - ヘルスチェックステータスの表示
  - インタラクティブボタンによる操作
- セキュリティ機能
  - チャンネル/ユーザー制限
  - コマンドインジェクション対策
  - 非rootユーザーでの実行（Dockerイメージ）
- 運用機能
  - 環境変数による設定
  - Docker操作のタイムアウト
  - 並行操作の制御
  - Docker Composeプロジェクト名の指定
  - マルチプラットフォーム対応（linux/amd64, linux/arm64）
- リリース機能
  - GitHub Actionsによる自動リリース
  - Docker イメージの自動ビルド・公開（ghcr.io）
  - GoReleaserによるクロスプラットフォームビルド

### Security
- サービス名の検証によるコマンドインジェクション対策
- Docker操作に適切なタイムアウト設定（操作により5秒〜30秒）
- Dockerイメージの非rootユーザー実行

[Unreleased]: https://github.com/hideA88/game-server-watchdog/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/hideA88/game-server-watchdog/releases/tag/v0.0.1
