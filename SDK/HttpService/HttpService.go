package HttpService

// -------------------------------------------------------------------------------------
import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Tools"
	"golang.org/x/crypto/pkcs12"
)

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
// HttpService 模擬 Java 的 HttpService 類別
type HttpService struct {
	_HttpServer  *http.Server
	_HttpsServer *http.Server
	_Mux         *http.ServeMux
	_MuxLock     sync.RWMutex

	_HttpPort      int
	_HttpsPort     int
	_SSLCert       string
	_SSLKeyFile    string
	_SSLPassword   string
	_PoolMethod    string
	_MaxConnection int

	_RootPath            string
	_DefaultHTML         string
	_CacheLock           sync.RWMutex
	_DefaultCacheControl string
	_EnableCache         bool

	_Handlers map[string]HttpAPI_Callback
}

// -------------------------------------------------------------------------------------
// NewHttpService 模擬 Java 建構子
func Create(_http_port, _https_port int, _ssl_cert, _ssl_key_file, _ssl_pwd string) *HttpService {
	_this := &HttpService{
		_HttpPort:    _http_port,
		_HttpsPort:   _https_port,
		_SSLCert:     _ssl_cert,
		_SSLKeyFile:  _ssl_key_file,
		_SSLPassword: _ssl_pwd,
		_Mux:         http.NewServeMux(),
		_Handlers:    make(map[string]HttpAPI_Callback),
	}

	_this._Mux.HandleFunc("/", _this.serveHTTP)
	_this.InitExecutor(false, "sync", 8, 800, 500)

	if _this._HttpPort > 0 {
		_this._HttpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", _this._HttpPort),
			Handler: _this._Mux,
		}
	}

	if _this._HttpsPort > 0 && _this._SSLCert != "" {
		_this._HttpsServer = &http.Server{
			Addr:      fmt.Sprintf(":%d", _this._HttpsPort),
			Handler:   _this._Mux,
			TLSConfig: secureServerTLSConfig(),
		}
	}

	return _this
}

func secureServerTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}
}

//-------------------------------------------------------------------------------------
// 設定方法
//-------------------------------------------------------------------------------------

func (_this *HttpService) SetRootPath(_path string) {
	_this._RootPath = _path
	if len(_this._RootPath) > 0 && strings.HasSuffix(_this._RootPath, "/") {
		_this._RootPath = _this._RootPath[:len(_this._RootPath)-1]
	}
}

// -------------------------------------------------------------------------------------
func (_this *HttpService) SetDefaultHTML(_default_html string) {
	_this._DefaultHTML = _default_html
}

// -------------------------------------------------------------------------------------
func (_this *HttpService) SetDefaultCacheControl(_control string) {
	_this._DefaultCacheControl = _control
}

// -------------------------------------------------------------------------------------
// InitExecutor 在 Go 中簡化實作 (因為 Go 自動處理 Goroutine 池)
func (_this *HttpService) InitExecutor(_force bool, _method string, _core, _max, _timeout int) {
	_this._PoolMethod = _method
	_this._MaxConnection = _max
	// Go 的 http.Server 會自動根據需求增長，這裡主要設定逾時限制
}

// -------------------------------------------------------------------------------------
// ServeHTTP 實作 http.Handler 介面 (對應 Java handle/Process)
func (_this *HttpService) serveHTTP(_w http.ResponseWriter, _r *http.Request) {

	_uriOrg, _ := url.PathUnescape(_r.RequestURI)

	if checkURI(_w, _uriOrg) == false {
		return
	}

	// _RootPath 沒設定就直接 404，避免空字串拼接 _uri 後從檔案系統根目錄送出任意檔案
	if strings.TrimSpace(_this._RootPath) == "" {
		http.Error(_w, "Not Found", http.StatusNotFound)
		return
	}

	_uri := strings.Split(_uriOrg, "?")[0]

	// 用 filepath.Join + Clean 標準化路徑後驗證仍位於 _RootPath 之下，防止符號連結 / 編碼變體繞過 ".."
	_absRoot, _err := filepath.Abs(_this._RootPath)
	if _err != nil {
		http.Error(_w, "Not Found", http.StatusNotFound)
		return
	}
	_absFile, _err := filepath.Abs(filepath.Join(_absRoot, filepath.FromSlash(_uri)))
	if _err != nil || (_absFile != _absRoot && !strings.HasPrefix(_absFile, _absRoot+string(filepath.Separator))) {
		http.Error(_w, "Not Acceptable", http.StatusNotAcceptable)
		return
	}
	_absFile, _err = resolvePublicStaticFile(_absRoot, _absFile, _this._DefaultHTML)
	if _err != nil {
		http.Error(_w, "Not Found", http.StatusNotFound)
		return
	}

	_cacheControl := Tools.If(_this._EnableCache, _this._DefaultCacheControl, "no-cache")

	// 視 client 的 Accept-Encoding 啟用 gzip / br；ServeFile 內部會 sniff Content-Type，
	// 圖片 / 影片等已壓縮格式由 wrapper 自動 skip，Range request 也已先在 MaybeCompressWriter 內 bypass
	_w, _release := MaybeCompressWriter(_w, _r)
	defer _release()

	_w.Header().Add("Cache-Control", _cacheControl.(string))
	http.ServeFile(_w, _r, _absFile)
}

func resolvePublicStaticFile(_absRoot string, _absFile string, _defaultHTML string) (string, error) {
	// Resolve symlinks before serving to ensure the requested file remains under
	// the configured public root; lexical path checks alone are insufficient.
	_realRoot, _err := filepath.EvalSymlinks(_absRoot)
	if _err != nil {
		return "", _err
	}
	_realFile, _err := filepath.EvalSymlinks(_absFile)
	if _err != nil || (_realFile != _realRoot && !strings.HasPrefix(_realFile, _realRoot+string(filepath.Separator))) {
		return "", os.ErrNotExist
	}
	_info, _err := os.Stat(_absFile)
	if _err != nil || !_info.Mode().IsRegular() && !_info.IsDir() {
		return "", os.ErrNotExist
	}
	if !_info.IsDir() {
		return _absFile, nil
	}

	_candidates := []string{}
	if _absFile == _absRoot && strings.TrimSpace(_defaultHTML) != "" {
		_candidates = append(_candidates, filepath.Base(strings.TrimSpace(_defaultHTML)))
	}
	_candidates = append(_candidates, "index.html")
	for _, _name := range _candidates {
		_candidate := filepath.Join(_absFile, _name)
		_realCandidate, _candidateErr := filepath.EvalSymlinks(_candidate)
		if _candidateErr != nil || (_realCandidate != _realRoot && !strings.HasPrefix(_realCandidate, _realRoot+string(filepath.Separator))) {
			continue
		}
		_candidateInfo, _candidateErr := os.Stat(_candidate)
		if _candidateErr == nil && _candidateInfo.Mode().IsRegular() {
			return _candidate, nil
		}
	}
	return "", os.ErrNotExist
}

// -------------------------------------------------------------------------------------
// CreateRestfulAPI 註冊路由處理器
func (_this *HttpService) AddRestfulAPI(_uri string, _callback HttpAPI_Callback) {

	if _callback != nil {

		_this._MuxLock.Lock()
		defer _this._MuxLock.Unlock()

		if strings.HasSuffix(_uri, "/") == false {
			_uri = _uri + "/"
		}
		if !strings.HasPrefix(_uri, "/") {
			_uri = "/" + _uri
		}
		// "/" 已由靜態檔案 dispatcher 佔用，避免重複註冊造成 ServeMux panic。
		if _uri == "/" {
			return
		}

		// ServeMux 不支援動態移除路由；每個 URI 只註冊一次 dispatcher。
		// 後續重複 Add 僅更新 callback，避免重複 HandleFunc 導致 panic。
		if _, exists := _this._Handlers[_uri]; !exists {
			_this._Mux.HandleFunc(_uri, func(w http.ResponseWriter, r *http.Request) {
				_this._MuxLock.RLock()
				_callback, exists := _this._Handlers[_uri]
				_this._MuxLock.RUnlock()
				if !exists || _callback == nil {
					http.NotFound(w, r)
					return
				}
				CreateHttpAPI(_callback).servHTTP(w, r)
			})
		}
		_this._Handlers[_uri] = _callback

		//Tools.ConsolePrint("AddRestfulAPI : " + _uri)
	}
}

// -------------------------------------------------------------------------------------
// RemoveRestfulAPI 移除路徑 (在標準 http.ServeMux 中較難實現，此處提供邏輯封裝)
func (_this *HttpService) RemoveRestfulAPI(_uri string) {

	_this._MuxLock.Lock()
	defer _this._MuxLock.Unlock()
	if strings.HasSuffix(_uri, "/") == false {
		_uri += "/"
	}
	if !strings.HasPrefix(_uri, "/") {
		_uri = "/" + _uri
	}
	// 保留已註冊標記，重新加入路由時不再重複註冊 ServeMux。
	if _, exists := _this._Handlers[_uri]; exists {
		_this._Handlers[_uri] = nil
	}
}

// -------------------------------------------------------------------------------------
// run 啟動服務 (對應 Java 的 start() -> run())
func (_this *HttpService) Run() {
	if err := _this.RunWithError(); err != nil {
		Tools.Log.Print(Tools.LL_Error, "HTTP startup: %v", err)
	}
}

// RunWithError 先完成所有 listener 與 TLS 憑證初始化，再啟動 Serve goroutine。
// 任一 listener 失敗都不會留下半啟動的 HTTP/HTTPS 服務。
func (_this *HttpService) RunWithError() error {
	var _httpListener, _httpsListener net.Listener
	var _err error
	if _this._HttpServer != nil {
		_httpListener, _err = net.Listen("tcp", _this._HttpServer.Addr)
		if _err != nil {
			return fmt.Errorf("HTTP listen %s: %w", _this._HttpServer.Addr, _err)
		}
	}
	if _this._HttpsServer != nil {
		_cert, _certErr := _this.loadTLSCertificate()
		if _certErr != nil {
			if _httpListener != nil {
				_httpListener.Close()
			}
			return fmt.Errorf("HTTPS TLS load: %w", _certErr)
		}
		_this._HttpsServer.TLSConfig.Certificates = []tls.Certificate{_cert}
		_httpsListener, _err = net.Listen("tcp", _this._HttpsServer.Addr)
		if _err != nil {
			if _httpListener != nil {
				_httpListener.Close()
			}
			return fmt.Errorf("HTTPS listen %s: %w", _this._HttpsServer.Addr, _err)
		}
		_httpsListener = tls.NewListener(_httpsListener, _this._HttpsServer.TLSConfig)
	}
	if _httpListener != nil {
		go func() {
			Tools.Log.Print(Tools.LL_Info, "Http Listen at : %d", _this._HttpPort)
			if _err := _this._HttpServer.Serve(_httpListener); _err != nil && _err != http.ErrServerClosed {
				Tools.Log.Print(Tools.LL_Error, "HTTP Listen Error: %v", _err)
			}
		}()
	}
	if _httpsListener != nil {
		go func() {
			Tools.Log.Print(Tools.LL_Info, "Https Listen at : %d", _this._HttpsPort)
			if _err := _this._HttpsServer.Serve(_httpsListener); _err != nil && _err != http.ErrServerClosed {
				Tools.Log.Print(Tools.LL_Error, "HTTPS Listen Error: %v", _err)
			}
		}()
	}
	return nil
}

// -------------------------------------------------------------------------------------
// Close 關閉伺服器
func (_this *HttpService) Close() bool {
	_ctx, _cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer _cancel()

	ok := true
	for _, server := range []*http.Server{_this._HttpServer, _this._HttpsServer} {
		if server != nil {
			if err := server.Shutdown(_ctx); err != nil {
				ok = false
				Tools.Log.Print(Tools.LL_Error, "HTTP shutdown: %v", err)
				server.Close()
			}
		}
	}
	return ok
}

// -------------------------------------------------------------------------------------
// GetHttpPort 取得實際通訊埠
func (_this *HttpService) GetHttpPort() int {
	return _this._HttpPort
}

// -------------------------------------------------------------------------------------
func (_this *HttpService) loadTLSCertificate() (tls.Certificate, error) {
	_ext := strings.ToLower(filepath.Ext(strings.TrimSpace(_this._SSLCert)))

	switch _ext {
	case ".p12", ".pfx":
		return _this.loadTLSCertificateFromP12()
	default:
		return _this.loadTLSCertificateFromPEM()
	}
}

// -------------------------------------------------------------------------------------
func (_this *HttpService) loadTLSCertificateFromPEM() (tls.Certificate, error) {
	if strings.TrimSpace(_this._SSLCert) == "" || strings.TrimSpace(_this._SSLKeyFile) == "" {
		return tls.Certificate{}, fmt.Errorf("ssl_key 與 ssl_key_file 必須同時設定，ssl_key_file 需對應 .key 檔案")
	}

	return tls.LoadX509KeyPair(_this._SSLCert, _this._SSLKeyFile)
}

// -------------------------------------------------------------------------------------
func (_this *HttpService) loadTLSCertificateFromP12() (tls.Certificate, error) {
	_p12Bytes, _err := os.ReadFile(_this._SSLCert)
	if _err != nil {
		return tls.Certificate{}, _err
	}

	_blocks, _err := pkcs12.ToPEM(_p12Bytes, _this._SSLPassword)
	if _err != nil {
		return tls.Certificate{}, _err
	}

	var _certPEM []byte
	var _keyPEM []byte

	for _, _block := range _blocks {
		if _block == nil {
			continue
		}

		_encoded := pem.EncodeToMemory(_block)
		if strings.Contains(_block.Type, "PRIVATE KEY") {
			_keyPEM = append(_keyPEM, _encoded...)
			continue
		}

		if strings.Contains(_block.Type, "CERTIFICATE") {
			_certPEM = append(_certPEM, _encoded...)
		}
	}

	if len(_certPEM) == 0 || len(_keyPEM) == 0 {
		return tls.Certificate{}, fmt.Errorf("ssl_key 指向的 p12 內容缺少憑證或私鑰")
	}

	return tls.X509KeyPair(_certPEM, _keyPEM)
}

// -------------------------------------------------------------------------------------
