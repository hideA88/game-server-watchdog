# リリース手順

## 準備

1. **GoReleaser のインストール**
   ```bash
   # macOS
   brew install goreleaser/tap/goreleaser

   # Linux
   echo 'deb [trusted=yes] https://repo.goreleaser.com/apt/ /' | sudo tee /etc/apt/sources.list.d/goreleaser.list
   sudo apt update
   sudo apt install goreleaser

   # または Go でインストール
   go install github.com/goreleaser/goreleaser@latest
   ```

2. **GitHub Personal Access Token の設定**
   - https://github.com/settings/tokens にアクセス
   - 「Generate new token (classic)」をクリック
   - `repo` スコープを選択
   - トークンを環境変数に設定：
     ```bash
     export GITHUB_TOKEN=your_token_here
     ```

## リリース方法

### 1. ローカルでのテスト

```bash
# ドライラン（実際にはリリースしない）
make release-dry-run

# ローカルビルドのテスト
make release-local
```

### 2. バージョンタグの作成

```bash
# セマンティックバージョニングに従う
# v1.0.0, v1.0.1, v1.1.0 など
make tag VERSION=v1.0.0

# タグをプッシュ
git push origin v1.0.0
```

### 3. 自動リリース

タグをプッシュすると、GitHub Actions が自動的に以下を実行します：
1. テストの実行
2. 各プラットフォーム向けバイナリのビルド
3. GitHub Releases へのアップロード
4. Docker イメージの作成とプッシュ（GitHub Container Registry）

### 4. リリースの確認

- GitHub Releases ページを確認：
  https://github.com/hideA88/game-server-watchdog/releases

- Docker イメージを確認：
  ```bash
  docker pull ghcr.io/hidea88/game-server-watchdog:v1.0.0
  ```

## バージョニングルール

セマンティックバージョニング（SemVer）に従います：

- **MAJOR** (v1.0.0 → v2.0.0): 破壊的変更
- **MINOR** (v1.0.0 → v1.1.0): 新機能の追加（後方互換性あり）
- **PATCH** (v1.0.0 → v1.0.1): バグ修正

## プレリリース

開発版をリリースする場合：

```bash
# ベータ版
make tag VERSION=v1.0.0-beta.1

# リリース候補
make tag VERSION=v1.0.0-rc.1
```

## トラブルシューティング

### GoReleaser が失敗する場合

1. Go のバージョンを確認（1.21以上が必要）
   ```bash
   go version
   ```

2. 依存関係を更新
   ```bash
   make deps
   ```

3. テストが通ることを確認
   ```bash
   make test
   ```

### Docker ビルドが失敗する場合

1. Docker Buildx が有効か確認
   ```bash
   docker buildx version
   ```

2. ローカルでビルドテスト
   ```bash
   make docker-build
   ```

## 手動リリース（緊急時）

GitHub Actions が使えない場合：

```bash
# 環境変数の設定
export GITHUB_TOKEN=your_token_here

# 手動でリリース
goreleaser release --clean
```