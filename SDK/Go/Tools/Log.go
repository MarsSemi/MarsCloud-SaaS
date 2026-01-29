package Tools

import (
	"fmt"
	"net"
	"os"
	"time"
)

// -------------------------------------------------------------------------------------
// 日誌系統 (簡化版 LogOut)
// -------------------------------------------------------------------------------------
var Log = createLogger()

// -------------------------------------------------------------------------------------
type LogLevel int

// -------------------------------------------------------------------------------------
const (
	LL_Debug LogLevel = iota
	LL_Normal
	LL_Info
	LL_Warning
	LL_Error
)

// -------------------------------------------------------------------------------------
type LogOut struct {
	_RemoteLoggerHost string
	_RemoteLoggerPort int
	_Log_ShowLevel    LogLevel
	_Log_ShowColor    bool
	_UDPConn          *net.UDPConn
}

// -------------------------------------------------------------------------------------
func createLogger() *LogOut {
	return &LogOut{
		_Log_ShowLevel:    LL_Debug,
		_Log_ShowColor:    true,
		_RemoteLoggerPort: 8729,
	}
}

// -------------------------------------------------------------------------------------
func (_l *LogOut) SetRemoteOutput(_enable bool, _addr string) {
	if !_enable {
		if _l._UDPConn != nil {
			_l._UDPConn.Close()
			_l._UDPConn = nil
		}
		return
	}
	_l._RemoteLoggerHost = _addr
	_raddr, _err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", _l._RemoteLoggerHost, _l._RemoteLoggerPort))
	if _err == nil {
		_l._UDPConn, _ = net.DialUDP("udp", nil, _raddr)
	}
}

// -------------------------------------------------------------------------------------
func (_l *LogOut) Print(_level LogLevel, _msgs ...string) {

	_msg := ""

	if _level < _l._Log_ShowLevel {
		return
	}

	for i := 0; i < len(_msgs); i++ {
		_msg += _msgs[i]
	}

	_timeStr := time.Now().Format("01/02 15:04:05")
	_levelStr := ""
	_colorCode := ""

	switch _level {
	case LL_Normal:
		_levelStr, _colorCode = "[Normal]", "\033[32m"
	case LL_Info:
		_levelStr, _colorCode = "[Info]", "\033[36m"
	case LL_Warning:
		_levelStr, _colorCode = "[Warn]", "\033[33m"
	case LL_Error:
		_levelStr, _colorCode = "[Error]", "\033[35m"
	default:
		_levelStr, _colorCode = "[Debug]", "\033[37m"
	}

	_out := fmt.Sprintf("[%s]%s %s", _timeStr, _levelStr, _msg)

	if _l._Log_ShowColor {
		fmt.Printf("%s%s\033[0m\n", _colorCode, _out)
	} else {
		fmt.Println(_out)
	}

	// 如果有設定遠端輸出，則發送 UDP 封包
	if _l._UDPConn != nil {
		_pid := os.Getpid()
		_remoteMsg := fmt.Sprintf("%d@%s", _pid, _out)
		_l._UDPConn.Write([]byte(_remoteMsg))
	}
}

// -------------------------------------------------------------------------------------
func (_l *LogOut) SetDisplayLevel(_level LogLevel) {
	_l._Log_ShowLevel = _level
}

// -------------------------------------------------------------------------------------
func ConsolePrint(_msg string) {
	Log.Print(LL_Info, _msg)
}
