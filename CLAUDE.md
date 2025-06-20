# Game Server Watchdog

このプロジェクトはゲームサーバーを監視・管理するDiscord botです。

## プロジェクト概要

- **言語**: Go
- **用途**: ゲームサーバーの監視、Discord経由での管理
- **主要ライブラリ**: discordgo

## 重要な設定

### 環境変数 (.env)
- `DISCORD_TOKEN`: Discord botトークン（必須）
- `DEBUG_MODE`: デバッグモード（true/false、デフォルト: false）
- `LOG_LEVEL`: ログレベル（debug/info/warn/error、**大文字小文字を区別しない**、デフォルト: info）
- `ALLOWED_CHANNEL_IDS`: 許可されたチャンネルID（カンマ区切り）
- `ALLOWED_USER_IDS`: 許可されたユーザーID（カンマ区切り）
- `DOCKER_COMPOSE_PATH`: docker-compose.ymlのパス（デフォルト: docker-compose.yml）
- `DOCKER_COMPOSE_PROJECT_NAME`: Docker Composeプロジェクト名（デフォルト: 空）

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

### ログシステム

#### 概要
構造化ロギングを採用し、context-basedでロガーを管理しています。

#### ログレベルの設定
優先順位（高い方が優先）：
1. `LOG_LEVEL`環境変数: `debug`, `info`, `warn`, `error`（**大文字小文字を区別しない**）
2. `DEBUG_MODE`環境変数: `true`の場合はDebugレベル、それ以外はInfoレベル

```bash
# デバッグモードで起動
DEBUG_MODE=true ./watchdog

# 特定のログレベルで起動（以下はすべて同じ動作）
LOG_LEVEL=warn ./watchdog
LOG_LEVEL=WARN ./watchdog
LOG_LEVEL=Warn ./watchdog
```

#### 使い方

```go
// 基本的な使い方（configから設定を渡す）
logger, err := logging.New(cfg.DebugMode, cfg.LogLevel)

// または、詳細な設定を指定して作成
logger, err := logging.NewWithConfig(&logging.Config{
    Level:       logging.InfoLevel,
    Development: false,
    Format:      "json",
})

// contextからロガーを取得
logger := logging.FromContext(ctx)

// ログ出力
logger.Info(ctx, "メッセージ", 
    logging.String("key", "value"),
    logging.Int("count", 42))

// エラーログ
logger.Error(ctx, "エラーが発生", logging.ErrorField(err))

// 新しいフィールドを追加
subLogger := logger.With(logging.String("component", "discord"))

// 名前付きロガー
namedLogger := logger.Named("subsystem")
```

#### contextへのロガー設定

```go
// ロガーをcontextに設定
ctx = logging.WithContext(ctx, logger)

// contextからロガーを取得（存在しない場合はnopLogger）
logger = logging.FromContext(ctx)
```

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