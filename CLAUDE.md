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
- `DOCKER_COMPOSE_PATH`: docker-compose.ymlのパス（デフォルト: docker-compose.yml）

### アクセス制限
- 特定のチャンネルIDのみでボットが反応するよう制限済み
- メンション必須：すべてのコマンドはボットをメンションして実行

## 実装済みコマンド

- `@bot ping`: 応答確認（pong!!を返す）
- `@bot help`: ヘルプメッセージを表示
- `@bot status`: サーバーのCPU使用率、メモリ使用量、ディスク空き容量を表示
- `@bot game-info`: Docker Composeで管理されているゲームサーバーの稼働状況を表示

## 開発時の注意事項

### コード品質管理
**コードの修正や追加を行った際は、必ず以下のコマンドを実行してください：**
1. `make format`: コードのフォーマット
2. `make test`: テストを実行（エラーが出た場合は必ず修正）
3. `make lint`: lintチェック（エラーが出た場合は必ず修正）

**重要**: コード修正後は必ず`make test`を実行し、全てのテストが通ることを確認してください。テストエラーが発生した場合は、必ず修正してからコミットしてください。

### コマンド実行
- `make format`: コードのフォーマット（goimports + go fmt）
- `make lint`: コードのlintチェック
- `make test`: 全テストを実行
- `make test-coverage`: カバレッジ付きでテスト実行
- `make coverage-filtered`: カバレッジ計測（.coverageignoreに記載されたファイルを除外）
- `go run cmd/watchdog/main.go`: ボット起動

### 新しいハンドラー追加時
1. `internal/bot/handler/`に新しいハンドラーファイルを作成
2. `handler.NewHandlers()`に追加
3. メンションチェックを必ず実装

### テストの書き方

**必ずテーブル駆動テスト（Table-Driven Tests）で実装すること**

テストはデフォルトでCPUコア数に応じて並列実行されます。並列実行に適したテストの場合は`t.Parallel()`を追加してください。

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name string
        // 入力パラメータ
        input string
        // 期待値
        want string
        wantErr bool
    }{
        {
            name: "正常系のケース",
            input: "test",
            want: "expected",
            wantErr: false,
        },
        {
            name: "異常系のケース",
            input: "",
            want: "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // テスト実行
            got, err := FunctionName(tt.input)
            
            // エラーチェック
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            // 結果チェック
            if got != tt.want {
                t.Errorf("FunctionName() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### カバレッジ計測
- `.coverageignore`ファイルにカバレッジから除外したいファイルパターンを記載
- 現在除外されているファイル：
  - `cmd/watchdog/main.go`: main関数
  - `internal/bot/bot.go`: 依存性の組み立てのみ
  - `**/mock*.go`: モックファイル
  - `**/generated*.go`: 生成されたファイル

### セキュリティ
- Discord tokenは絶対にコミットしない
- `.env`ファイルは`.gitignore`に含める

## 今後の実装予定
- サーバー監視機能（CPU、メモリ、ディスク使用率）
- ゲームサーバーの起動/停止制御
- Minecraft RCON対応（既に基盤あり）
- アラート通知機能