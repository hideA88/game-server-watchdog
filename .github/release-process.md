# リリースプロセス

## 定期リリースサイクル

### 1. 開発フロー
```
main ブランチ
  ↓
feature/* ブランチで開発
  ↓
Pull Request でレビュー
  ↓
main にマージ
  ↓
リリース準備
```

### 2. リリース前チェックリスト

- [ ] すべてのテストが通る（`make test`）
- [ ] lint エラーがない（`make lint`）
- [ ] CHANGELOG.md を更新
- [ ] バージョン番号を決定
- [ ] リリースノートを準備

### 3. バージョニングルール

```
v[MAJOR].[MINOR].[PATCH]

MAJOR: 破壊的変更
MINOR: 新機能（後方互換性あり）
PATCH: バグ修正
```

例：
- `v0.1.0` → `v0.2.0` (新機能追加)
- `v0.2.0` → `v0.2.1` (バグ修正)
- `v0.2.1` → `v1.0.0` (破壊的変更)

### 4. リリースタイプ

#### 通常リリース
```bash
make tag VERSION=v0.2.0
git push origin v0.2.0
```

#### プレリリース
```bash
# アルファ版
make tag VERSION=v0.2.0-alpha.1

# ベータ版
make tag VERSION=v0.2.0-beta.1

# リリース候補
make tag VERSION=v0.2.0-rc.1
```

#### ホットフィックス
```bash
# main から直接
git checkout main
git pull origin main
# 修正をコミット
make tag VERSION=v0.1.1
git push origin v0.1.1
```

### 5. リリース後のタスク

- [ ] リリースノートを確認
- [ ] Docker イメージが公開されているか確認
- [ ] ダウンロードリンクをテスト
- [ ] README.md のバージョン情報を更新
- [ ] 次のマイルストーンを作成

### 6. 自動化されているタスク

以下は GitHub Actions で自動実行：
- ✅ テストの実行
- ✅ 各プラットフォーム向けビルド
- ✅ GitHub Release の作成
- ✅ Docker イメージのプッシュ
- ✅ チェックサムの生成

### 7. リリースノートテンプレート

```markdown
## 🎉 新機能
- 機能A の追加 (#PR番号)
- 機能B の改善 (#PR番号)

## 🐛 バグ修正
- 問題X を修正 (#Issue番号)

## 🔧 改善
- パフォーマンスの向上
- コードのリファクタリング

## ⚠️ 破壊的変更
- API の変更点

## 📦 依存関係
- ライブラリX を v1.2.3 に更新

## 🙏 貢献者
@username1, @username2
```