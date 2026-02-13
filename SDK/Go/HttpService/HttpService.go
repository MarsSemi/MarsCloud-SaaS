package HttpService

// -------------------------------------------------------------------------------------
import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
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
	_SSLKey        string
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
func Create(_http_port, _https_port int, _ssl_key, _ssl_pwd string) *HttpService {
	_this := &HttpService{
		_HttpPort:    _http_port,
		_HttpsPort:   _https_port,
		_SSLKey:      _ssl_key,
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

	if _this._HttpsPort > 0 && _this._SSLKey != "" {
		// 這裡假設 SSL Key 是 PEM 格式路徑，或是使用先前 Security.go 實作的加載方式
		_this._HttpsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", _this._HttpsPort),
			Handler: _this._Mux,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		}
	}

	return _this
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

	//Tools.Log.Print(Tools.LL_Info, "Call Root API : "+_uriOrg)

	_uri := strings.Split(_uriOrg, "?")[0]
	_fn := _this._RootPath + _uri
	_cacheControl := Tools.If(_this._EnableCache, _this._DefaultCacheControl, "no-cache")

	if len(_fn) > 0 {

		_w.Header().Add("Cache-Control", _cacheControl.(string))

		http.ServeFile(_w, _r, _fn)
		return
	}

	http.Error(_w, "Not Found", http.StatusNotFound)
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

		_api := CreateHttpAPI(_callback)

		_this._Handlers[_uri] = _api.callBack
		_this._Mux.HandleFunc(_uri, _api.servHTTP)

		//Tools.ConsolePrint("AddRestfulAPI : " + _uri)
	}
}

// -------------------------------------------------------------------------------------
// RemoveRestfulAPI 移除路徑 (在標準 http.ServeMux 中較難實現，此處提供邏輯封裝)
func (_this *HttpService) RemoveRestfulAPI(_uri string) {

	_this._MuxLock.Lock()
	defer _this._MuxLock.Unlock()
	delete(_this._Handlers, _uri)

	http.HandleFunc(_uri, nil)
}

// -------------------------------------------------------------------------------------
// run 啟動服務 (對應 Java 的 start() -> run())
func (_this *HttpService) Run() {

	// HTTP Server
	if _this._HttpServer != nil {

		go func() {

			Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Http Listen at : %d", _this._HttpPort))

			if _err := _this._HttpServer.ListenAndServe(); _err != nil && _err != http.ErrServerClosed {

				Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("HTTP Listen Error: %v", _err))

			}
		}()
	}

	// HTTPS Server
	if _this._HttpsServer != nil {

		go func() {

			Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Https Listen at : %d", _this._HttpsPort))
			Tools.ConsolePrint(_this._SSLKey)
			Tools.ConsolePrint(_this._SSLPassword)
			// Go 內建 ListenAndServeTLS 需要 Cert 與 Key 檔案路徑
			if _err := _this._HttpsServer.ListenAndServeTLS(_this._SSLKey, _this._SSLPassword); _err != nil && _err != http.ErrServerClosed {

				Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("HTTPS Listen Error: %v\n", _err))

			}
		}()
	}
}

// -------------------------------------------------------------------------------------
// Close 關閉伺服器
func (_this *HttpService) Close() bool {
	_ctx, _cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer _cancel()

	if _this._HttpServer != nil {
		_this._HttpServer.Shutdown(_ctx)
	}
	if _this._HttpsServer != nil {
		_this._HttpsServer.Shutdown(_ctx)
	}
	return true
}

// -------------------------------------------------------------------------------------
// GetHttpPort 取得實際通訊埠
func (_this *HttpService) GetHttpPort() int {
	return _this._HttpPort
}

// -------------------------------------------------------------------------------------
