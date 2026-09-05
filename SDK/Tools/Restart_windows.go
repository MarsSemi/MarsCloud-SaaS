package Tools

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

const restartParentHandleEnv = "MARS_SERVICE_RESTART_PARENT_HANDLE"

// 子程序在應用程式初始化前等候原程序退出，避免搶綁尚未釋放的連接埠。
// 繼承程序 handle 而非輪詢 PID，避免 PID 重用與父程序提早退出的競爭。
func init() {
	_raw := os.Getenv(restartParentHandleEnv)
	if _raw == "" {
		return
	}
	os.Unsetenv(restartParentHandleEnv)
	_value, _err := strconv.ParseUint(_raw, 10, 64)
	if _err != nil || _value == 0 {
		fmt.Fprintln(os.Stderr, "重啟交接失敗：無效的父程序 handle")
		os.Exit(1)
	}
	_handle := windows.Handle(_value)
	_result, _err := windows.WaitForSingleObject(_handle, windows.INFINITE)
	windows.CloseHandle(_handle)
	if _err != nil || _result != windows.WAIT_OBJECT_0 {
		fmt.Fprintf(os.Stderr, "重啟交接失敗：等待父程序退出，結果 %d，錯誤 %v\n", _result, _err)
		os.Exit(1)
	}
}

func restartProcess(_self string, _args, _env []string) error {
	_parent, _err := windows.OpenProcess(windows.SYNCHRONIZE, true, uint32(os.Getpid()))
	if _err != nil {
		return fmt.Errorf("建立重啟交接 handle 失敗: %w", _err)
	}
	defer windows.CloseHandle(_parent)

	_cmd := exec.Command(_self, _args[1:]...)
	_cmd.Stdin, _cmd.Stdout, _cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	for _, _entry := range _env {
		_key, _, _ := strings.Cut(_entry, "=")
		if !strings.EqualFold(_key, restartParentHandleEnv) {
			_cmd.Env = append(_cmd.Env, _entry)
		}
	}
	_cmd.Env = append(_cmd.Env, restartParentHandleEnv+"="+strconv.FormatUint(uint64(_parent), 10))
	_cmd.SysProcAttr = &syscall.SysProcAttr{
		AdditionalInheritedHandles: []syscall.Handle{syscall.Handle(_parent)},
	}
	if _err := _cmd.Start(); _err != nil {
		return fmt.Errorf("啟動重啟子程序失敗: %w", _err)
	}
	// 子程序已由作業系統建立；原程序退出前，子程序仍停在交接屏障。
	_ = _cmd.Process.Release()
	return nil
}
