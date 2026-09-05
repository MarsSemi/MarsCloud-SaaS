//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris && !windows

package Tools

import (
	"fmt"
	"runtime"
)

func restartProcess(_self string, _args, _env []string) error {
	return fmt.Errorf("平台 %s 不支援內建程序重啟，請使用外部服務管理器", runtime.GOOS)
}
