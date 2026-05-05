// Package usermsg は Discord 上で表示するユーザー向けメッセージの生成を提供します。
package usermsg

// DockerPermissionMessage は Docker socket への権限エラー発生時に
// Discord で表示するユーザー向けメッセージを返す。
// ホスト実行とコンテナ実行の両方の対処方法を案内する。
func DockerPermissionMessage() string {
	return `Docker権限エラーが発生しました。Docker socketへのアクセス権限がありません。

**ホスト上で直接実行している場合**
1. dockerグループに属しているか確認: ` + "`groups`" + `
2. 追加して再ログイン: ` + "`sudo usermod -aG docker $USER`" + ` 後に再ログイン
3. 一時的な反映: ` + "`newgrp docker`" + `

**Docker コンテナ内で実行している場合**
1. docker-entrypoint.sh の起動ログを確認 (` + "`docker logs <container>`" + `)
2. compose.yml で ` + "`group_add`" + ` にホストの docker GID を指定
   （` + "`getent group docker`" + ` で確認）
3. ` + "`/var/run/docker.sock`" + ` がコンテナにマウントされているか確認

詳細は管理者にお問い合わせください。`
}
