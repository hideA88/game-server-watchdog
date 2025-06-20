//go:build linux

package system

import "syscall"

// syscallStatfs はLinux用のstatfs構造体
type syscallStatfs syscall.Statfs_t

// syscallStatfsFunc はLinux用のstatfs関数
func syscallStatfsFunc(path string, stat *syscallStatfs) error {
	return syscall.Statfs(path, (*syscall.Statfs_t)(stat))
}
