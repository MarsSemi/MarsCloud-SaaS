package HttpService

// -------------------------------------------------------------------------------------
import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

// -------------------------------------------------------------------------------------
// 不應再壓縮的 Content-Type 前綴（已是壓縮 / 二進位格式，再壓縮反而吃 CPU 沒收益）
var _noCompressContentTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"video/",
	"audio/",
	"application/zip",
	"application/gzip",
	"application/x-gzip",
	"application/x-bzip2",
	"application/x-xz",
	"application/x-brotli",
	"application/vnd.brotli",
	"application/octet-stream",
}

// -------------------------------------------------------------------------------------
// 壓縮 writer pool，重用以減少 GC 壓力（每次 Reset 即可指向新的 underlying writer）
var _gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

var _brotliPool = sync.Pool{
	New: func() interface{} {
		return brotli.NewWriter(io.Discard)
	},
}

// -------------------------------------------------------------------------------------
// compressResponseWriter 包裝 http.ResponseWriter 提供透明 gzip / br 壓縮
type compressResponseWriter struct {
	http.ResponseWriter
	_cw         io.WriteCloser
	_encoding   string
	_headerSent bool
	_skip       bool // Content-Type 屬於不壓縮類型 / 上游已自行設 Content-Encoding 時為 true
}

// -------------------------------------------------------------------------------------
func (_w *compressResponseWriter) WriteHeader(_statusCode int) {
	if _w._headerSent {
		return
	}
	_w._headerSent = true

	// 1xx / 204 / 304 規格上不能有 body，多寫壓縮 header/trailer 會讓 client 視為協定錯誤
	if _statusCode == http.StatusNoContent ||
		_statusCode == http.StatusNotModified ||
		(_statusCode >= 100 && _statusCode < 200) {
		_w._skip = true
		_w.ResponseWriter.WriteHeader(_statusCode)
		return
	}

	_h := _w.ResponseWriter.Header()

	// 上游已自設 Content-Encoding（例如本身就是 gzip 過的內容）就不重複壓縮
	if _h.Get("Content-Encoding") != "" {
		_w._skip = true
		_w.ResponseWriter.WriteHeader(_statusCode)
		return
	}

	if shouldSkipCompression(_h.Get("Content-Type")) {
		_w._skip = true
		_w.ResponseWriter.WriteHeader(_statusCode)
		return
	}

	_h.Set("Content-Encoding", _w._encoding)
	_h.Add("Vary", "Accept-Encoding")
	_h.Del("Content-Length") // 壓縮後長度會變，交由 chunked transfer 處理
	_w.ResponseWriter.WriteHeader(_statusCode)
}

// -------------------------------------------------------------------------------------
func (_w *compressResponseWriter) Write(_data []byte) (int, error) {
	if !_w._headerSent {
		// Content-Type 沒設時先嗅探，避免 Go 後續 implicit 設成 text/plain 把判斷蓋掉
		if _w.ResponseWriter.Header().Get("Content-Type") == "" {
			_w.ResponseWriter.Header().Set("Content-Type", http.DetectContentType(_data))
		}
		_w.WriteHeader(http.StatusOK)
	}
	if _w._skip {
		return _w.ResponseWriter.Write(_data)
	}
	return _w._cw.Write(_data)
}

// -------------------------------------------------------------------------------------
// Flush 實作 http.Flusher，避免包裝後破壞 SSE / chunked streaming
func (_w *compressResponseWriter) Flush() {
	if _w._cw != nil && !_w._skip {
		if _f, _ok := _w._cw.(interface{ Flush() error }); _ok {
			_ = _f.Flush()
		}
	}
	if _f, _ok := _w.ResponseWriter.(http.Flusher); _ok {
		_f.Flush()
	}
}

// -------------------------------------------------------------------------------------
// MaybeCompressWriter 視 request 條件回傳 gzip / br 包裝過的 writer 與 release 函式
// release 必須以 defer 呼叫以 flush 殘餘資料並把壓縮 writer 歸還 pool
// 不需壓縮的情境（client 不接受支援的 encoding / Range request）會回傳原 writer 與 noop release
func MaybeCompressWriter(_w http.ResponseWriter, _r *http.Request) (http.ResponseWriter, func()) {
	// Range request 是針對未壓縮資料的 byte 範圍，壓縮後 byte offset 完全失準
	if _r.Header.Get("Range") != "" {
		return _w, func() {}
	}

	_encoding := preferredCompression(_r)
	if _encoding == "" {
		return _w, func() {}
	}

	if _encoding == "br" {
		_bw := _brotliPool.Get().(*brotli.Writer)
		_bw.Reset(_w)
		_wrapped := &compressResponseWriter{
			ResponseWriter: _w,
			_cw:            _bw,
			_encoding:      _encoding,
		}
		return _wrapped, func() {
			if !_wrapped._skip {
				_ = _bw.Close()
			}
			_bw.Reset(io.Discard)
			_brotliPool.Put(_bw)
		}
	}

	_gw := _gzipPool.Get().(*gzip.Writer)
	_gw.Reset(_w)
	_wrapped := &compressResponseWriter{
		ResponseWriter: _w,
		_cw:            _gw,
		_encoding:      _encoding,
	}
	return _wrapped, func() {
		if !_wrapped._skip {
			_ = _gw.Close()
		}
		_gw.Reset(io.Discard)
		_gzipPool.Put(_gw)
	}
}

// -------------------------------------------------------------------------------------
// MaybeGzipWriter 保留舊 API 名稱相容性；實際會依 Accept-Encoding 選擇 br 或 gzip。
func MaybeGzipWriter(_w http.ResponseWriter, _r *http.Request) (http.ResponseWriter, func()) {
	return MaybeCompressWriter(_w, _r)
}

// -------------------------------------------------------------------------------------
func preferredCompression(_r *http.Request) string {
	_header := _r.Header.Get("Accept-Encoding")
	if strings.TrimSpace(_header) == "" {
		return ""
	}

	_qByEncoding := map[string]float64{}
	_wildcardQ := -1.0
	for _, _item := range strings.Split(_header, ",") {
		_encoding, _q, _ok := parseAcceptEncoding(_item)
		if !_ok {
			continue
		}
		if _encoding == "*" {
			_wildcardQ = _q
			continue
		}
		_qByEncoding[_encoding] = _q
	}

	// 同 q 值時優先 br；若 client 以 q 明確指定 gzip 較高，則尊重 client 偏好。
	_supported := []string{"br", "gzip"}
	_bestEncoding := ""
	_bestQ := 0.0
	for _, _encoding := range _supported {
		_q, _exists := _qByEncoding[_encoding]
		if !_exists {
			_q = _wildcardQ
		}
		if _q <= 0 {
			continue
		}
		if _bestEncoding == "" || _q > _bestQ {
			_bestEncoding = _encoding
			_bestQ = _q
		}
	}

	return _bestEncoding
}

// -------------------------------------------------------------------------------------
func parseAcceptEncoding(_item string) (string, float64, bool) {
	_parts := strings.Split(_item, ";")
	_encoding := strings.ToLower(strings.TrimSpace(_parts[0]))
	if _encoding == "" {
		return "", 0, false
	}

	_q := 1.0
	for _, _part := range _parts[1:] {
		_param := strings.SplitN(strings.TrimSpace(_part), "=", 2)
		if len(_param) != 2 || strings.ToLower(strings.TrimSpace(_param[0])) != "q" {
			continue
		}
		_parsedQ, _err := strconv.ParseFloat(strings.TrimSpace(_param[1]), 64)
		if _err != nil {
			return "", 0, false
		}
		if _parsedQ < 0 {
			_parsedQ = 0
		}
		if _parsedQ > 1 {
			_parsedQ = 1
		}
		_q = _parsedQ
		break
	}

	return _encoding, _q, true
}

// -------------------------------------------------------------------------------------
func shouldSkipCompression(_ct string) bool {
	_ct = strings.ToLower(strings.TrimSpace(_ct))
	if _ct == "" {
		return false
	}
	// 去除 ";charset=..." 等參數，只比對 type/subtype
	if _idx := strings.Index(_ct, ";"); _idx >= 0 {
		_ct = strings.TrimSpace(_ct[:_idx])
	}
	for _, _prefix := range _noCompressContentTypes {
		if strings.HasPrefix(_ct, _prefix) {
			return true
		}
	}
	return false
}

// -------------------------------------------------------------------------------------
