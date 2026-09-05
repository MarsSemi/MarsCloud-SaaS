//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package Tools

import "syscall"

// 直接替換程序，保留 PID、工作目錄、標準輸入輸出與服務管理器的追蹤關係。
// 不先清除舊服務資源，exec 失敗時才能完整保留目前服務。
func restartProcess(_self string, _args, _env []string) error {
	return syscall.Exec(_self, _args, _env)
}
