package MarsService

//-------------------------------------------------------------------------------------
import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/AsyncTaskProcessor"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/HttpService"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/MQTTClient"
	MarsMQTTServer "github.com/MarsSemi/MarsCloud-SaaS/SDK/MQTTServer"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/MarsClient"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Tools"
)

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
type IMarsService interface {
	OnMQTTConnected()
	OnMQTTMessage(_topic string, _payload string)
	OnMQTTConnectionLost(_err error)
	OnPropertyChange(*MarsJSON.JSONObject)

	BeforeServiceStop()
	Process()
}

// -------------------------------------------------------------------------------------
// Service CALLBACK
// -------------------------------------------------------------------------------------
type serviceCallback struct {
	service *MarsService
}

// -------------------------------------------------------------------------------------
func (_this *serviceCallback) OnConnected() {
	Tools.Log.Print(Tools.LL_Info, "MQTT connected")
	_this.service.impl.OnMQTTConnected()
}

// -------------------------------------------------------------------------------------
func (_this *serviceCallback) OnConnectionLost(_err error) {
	Tools.Log.Print(Tools.LL_Warning, "MQTT connection lost")
	_this.service.impl.OnMQTTConnectionLost(_err)
	_this.service.ResetMQTTClient(_this.service.MQTT_Topic)
}

// -------------------------------------------------------------------------------------
func (_this *serviceCallback) OnDeliveryComplete(_token string) {}

// -------------------------------------------------------------------------------------
func (_this *serviceCallback) OnMessageArrived(_topic string, _msg *MQTTClient.MQTTMessage) {

	defer func() {
		if _r := recover(); _r != nil {
			Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("MQTT Message Error: %v", _r))
		}
	}()

	_payload := string(_msg.GetPayload())

	if _this.service.AsyncTaskProcessor != nil {

		if _topic == _this.service.MQTT_AsyncTask_Topic {
			_this.service.AsyncTaskProcessor.OnMQTTMessage(_topic, _payload)
		} else {

			_this.service.onMQTTDefault(_topic, _payload)
			_this.service.impl.OnMQTTMessage(_topic, _payload)
		}
	}
}

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
type MarsService struct {
	MarsClient         *MarsClient.MarsClient
	MQTTClient         *MQTTClient.MQTTClient
	LocalMQTTServer    *MarsMQTTServer.MQTTServer
	HttpService        *HttpService.HttpService
	Property           *MarsJSON.JSONObject
	ServiceInfo        *MarsJSON.JSONObject
	AsyncTaskProcessor *AsyncTaskProcessor.AsyncTaskProcessor

	ServiceID        string
	ServiceName      string
	ServiceType      string
	ServiceVersion   string
	PropertyFileName string

	MQTT_Default_Topic   string
	MQTT_AsyncTask_Topic string
	MQTT_Topic           string
	MQTTCallback         MQTTClient.MQTTCallback

	defaultHttpPort  int
	defaultHttpsPort int
	ssl_Cert_File    string
	ssl_Key_File     string
	ssl_Key_Password string

	account                  string
	password                 string
	webHook                  string
	localMQTTMessageCallback MarsMQTTServer.MessageCallback

	SystemStartTime      int64
	IsDebugData          bool
	RestartAfterConflict bool

	syncTimer    *time.Ticker
	defaultTimer *time.Ticker // 用於 AutoGC

	stopChan    chan struct{}
	stopOnce    sync.Once // 整段停止清理只執行一次
	startOnce   sync.Once
	lifecycleMu sync.Mutex // 序列化啟動、停止與重啟，重複請求不排隊執行

	autoRestartTime    MarsJSON.JSONArray
	autoRestartMinutes []string       // restart_time 解析後的 "HH:MM" 清單，無效項目會被丟棄
	restartLocation    *time.Location // restart_time 比對所用時區，預設 time.Local
	registered         bool           // 服務是否已成功完成首次註冊（含 properties 上傳）
	propertyMu         sync.RWMutex   // 保護 Property 在 heartbeat 讀取與 HTTP /system/update_setting 寫入之間的並行存取
	impl               IMarsService   // 指向具體的實作物件_
}

// -------------------------------------------------------------------------------------
func Create(_propertyFileName string, _impl IMarsService) *MarsService {
	_this := &MarsService{
		PropertyFileName: _propertyFileName,
		SystemStartTime:  time.Now().UnixMilli(),
		ServiceInfo:      MarsJSON.NewJSONObject(""),
		stopChan:         make(chan struct{}),
		impl:             _impl,
	}

	_this.initCloseHook()
	_this.init(_propertyFileName)

	return _this
}

//-------------------------------------------------------------------------------------
// 核心初始化邏輯
//-------------------------------------------------------------------------------------

func (_this *MarsService) init(_propertyFileName string) {

	_this.Property = MarsJSON.NewJSONObject(Tools.File2String(_propertyFileName))

	// 初始化基本變數
	_this.defaultHttpPort = _this.Property.OptInt("http_port", 80)
	_this.defaultHttpsPort = _this.Property.OptInt("https_port", 433)
	_this.ssl_Cert_File = _this.Property.OptString("ssl_key", "")
	_this.ssl_Key_File = _this.Property.OptString("ssl_key_file", "")
	_this.ssl_Key_Password = _this.Property.OptString("ssl_key_password", "")

	_this.ServiceName = _this.Property.OptString("service_name", "Unknown Service")
	_this.ServiceID = fmt.Sprintf("%s-%d", Tools.GetMachineID(), _this.defaultHttpPort)

	_this.account = _this.Property.OptString("mars_cloud_account", "")
	_this.password = _this.Property.OptString("mars_cloud_password", "")
	_this.webHook = _this.getWebHook()
	_this.autoRestartTime = *_this.Property.OptJSONArray("restart_time") //["6:00:00", "14:12:24"]
	_this.autoRestartMinutes = parseRestartMinutes(&_this.autoRestartTime)
	_this.restartLocation = resolveRestartLocation(_this.Property.OptString("restart_timezone", ""))

	// 預設仍維持舊行為（略過 TLS 驗證），明確設成 false 才會啟用憑證驗證；同時影響 HttpPost 與 MQTT 連線
	Tools.DefaultInsecureTLS = _this.Property.OptBoolean("tls_skip_verify", true)

	_this.ResetWebService()

	fmt.Printf("\n------------------------------------\n")
	fmt.Printf("\n %s \n", _this.ServiceName)
	fmt.Printf("\n------------------------------------\n\n")

	Tools.Log.Print(Tools.LL_Info, "Service ID : %s", _this.ServiceID)
	Tools.Log.Print(Tools.LL_Info, "Process ID : %d", Tools.GetPID(nil))

	// 容器若 /etc/localtime 沒設會 fallback UTC，明確輸出有助於排查 restart_time 沒按預期觸發的問題
	_zoneName, _offset := time.Now().In(_this.restartLocation).Zone()
	Tools.Log.Print(Tools.LL_Info, "Restart Timezone : %s (%s, UTC%+d)", _this.restartLocation.String(), _zoneName, _offset/3600)
}

// -------------------------------------------------------------------------------------
// parseRestartMinutes 把 restart_time 各種輸入格式（"6:00", "06:00", "6:00:00" 等）統一成 "HH:MM"
// 解析失敗的項目會被丟棄並記 warning，避免 _targetTime[:5] 直接切片造成靜默 miss
func parseRestartMinutes(_arr *MarsJSON.JSONArray) []string {
	_result := make([]string, 0, _arr.Length())
	for _i := 0; _i < _arr.Length(); _i++ {
		_raw := strings.TrimSpace(_arr.OptString(_i, ""))
		if _raw == "" {
			continue
		}

		_parts := strings.Split(_raw, ":")
		if len(_parts) < 2 {
			Tools.Log.Print(Tools.LL_Warning, "Invalid restart_time entry %q, skipped", _raw)
			continue
		}
		_h, _err1 := strconv.Atoi(strings.TrimSpace(_parts[0]))
		_m, _err2 := strconv.Atoi(strings.TrimSpace(_parts[1]))
		if _err1 != nil || _err2 != nil || _h < 0 || _h > 23 || _m < 0 || _m > 59 {
			Tools.Log.Print(Tools.LL_Warning, "Invalid restart_time entry %q, skipped", _raw)
			continue
		}
		_result = append(_result, fmt.Sprintf("%02d:%02d", _h, _m))
	}
	return _result
}

// -------------------------------------------------------------------------------------
func resolveRestartLocation(_name string) *time.Location {
	_name = strings.TrimSpace(_name)
	if _name == "" {
		return time.Local
	}
	if _loc, _err := time.LoadLocation(_name); _err == nil {
		return _loc
	}
	Tools.Log.Print(Tools.LL_Warning, "Invalid restart_timezone %q, fallback to local", _name)
	return time.Local
}

// -------------------------------------------------------------------------------------
// Start 非同步啟動；以實際綁定 listener 的結果判斷成功，不依程序名稱或連接埠殺程序。
func (_this *MarsService) Start() {
	_this.startOnce.Do(func() {
		go func() {
			_this.lifecycleMu.Lock()
			defer _this.lifecycleMu.Unlock()
			if _this.isStopping() {
				return
			}

			_url := _this.Property.OptString("mars_cloud_url", "")
			_proj := _this.Property.OptString("mars_cloud_proj", "")
			_hasCloudConfig := _this.hasCompleteMarsCloudConfig(_url)

			if _this.HttpService != nil {
				_this.HttpService.SetRootPath(_this.Property.OptString("web_path", "./website"))
				_this.HttpService.SetDefaultCacheControl("public, max-age=43200")
				if _err := _this.HttpService.RunWithError(); _err != nil {
					Tools.Log.Print(Tools.LL_Error, "服務啟動失敗: %v", _err)
					os.Exit(1)
				}
			}
			if _this.shouldStartLocalMQTTServer(_hasCloudConfig) {
				if _err := _this.startLocalMQTTServer(); _err != nil {
					Tools.Log.Print(Tools.LL_Error, "服務啟動失敗: %v", _err)
					_this.CloseNetService()
					os.Exit(1)
				}
			}

			if _hasCloudConfig {
				go _this.connectCloudInBackground(_url, _proj)
			} else {
				Tools.Log.Print(Tools.LL_Info, "MarsCloud disabled: mars_cloud_url/account/password 缺少設定，使用一般 Server 模式啟動")
			}
			_this.ResetAutoRestart()
			_this.ResetAutoGC()
			Tools.Log.Print(Tools.LL_Info, "Service Start : %s %s", _this.ServiceName, _this.ServiceVersion)
		}()
	})
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) hasCompleteMarsCloudConfig(_url string) bool {
	return strings.TrimSpace(_url) != "" &&
		strings.TrimSpace(_this.account) != "" &&
		strings.TrimSpace(_this.password) != ""
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) shouldStartLocalMQTTServer(_hasCloudConfig bool) bool {
	return _this.Property.OptBoolean("mqtt_server_enable", false)
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) startLocalMQTTServer() error {
	if _this.LocalMQTTServer != nil {
		return nil
	}

	_config := MarsMQTTServer.Config{
		Host:           _this.Property.OptString("mqtt_bind", ""),
		TCPPort:        _this.Property.OptInt("mqtt_tcp_port", 1883),
		WSPort:         _this.Property.OptInt("mqtt_ws_port", 1884),
		SSLPort:        _this.Property.OptInt("mqtt_ssl_port", 8883),
		WSSPort:        _this.Property.OptInt("mqtt_wss_port", 8884),
		CertFile:       _this.Property.OptString("mqtt_tls_cert", ""),
		KeyFile:        _this.Property.OptString("mqtt_tls_key", ""),
		AllowAnonymous: _this.Property.OptBoolean("mqtt_allow_anonymous", false),
		Username:       _this.Property.OptString("mqtt_username", ""),
		Password:       _this.Property.OptString("mqtt_password", ""),
		OnMessage:      _this.localMQTTMessageCallback,
	}

	_this.LocalMQTTServer = MarsMQTTServer.Create(_config)

	if _err := _this.LocalMQTTServer.Start(); _err != nil {
		Tools.Log.Print(Tools.LL_Error, "Local MQTT server start fail: %s", _err.Error())
		_this.LocalMQTTServer = nil
		return _err
	}
	return nil
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) SetLocalMQTTMessageCallback(_callback MarsMQTTServer.MessageCallback) {
	_this.localMQTTMessageCallback = _callback
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) initCloseHook() {
	// 建立一個頻道來接收作業系統訊號
	_sigChan := make(chan os.Signal, 1)

	// 註冊要監聽的訊號 (SIGINT = Ctrl+C, SIGTERM = 停止訊號)
	signal.Notify(_sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 開啟一個 Goroutine 在背景等待訊號
	go func() {
		_sig := <-_sigChan // 這會阻塞直到收到訊號

		Tools.Log.Print(Tools.LL_Info, "- ")
		Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Get Closing Signal : %v, clean up ...", _sig))

		_this.ShutdownService()
	}()
}

// -------------------------------------------------------------------------------------
// connectCloudInBackground 在獨立 goroutine 完成雲端登入、MQTT 連線與首次註冊
// 設計理由：initMarsClient 內含無限重試 loop，若放在 startup 路徑會導致雲端不可達時整個 HTTP server 也無法啟動
func (_this *MarsService) connectCloudInBackground(_url string, _proj string) {
	defer Tools.GlobalRecovery()

	_this.initMarsClient(_url, _this.account, _this.password, _proj)
	if _this.MarsClient == nil || _this.isStopping() {
		return
	}

	_this.initMQTTClient(_this.MarsClient.GetServerURL())
	_this.AsyncTaskProcessor = AsyncTaskProcessor.Create(_this.MarsClient, _this.webHook)
	_this.doRegistry(true)
}

// -------------------------------------------------------------------------------------
// InitMarsClient 初始化與 MarsCloud 的連線
func (_this *MarsService) initMarsClient(_url, _account, _pass, _proj string) {
	if _url == "" || _account == "" || _pass == "" {
		return
	}

	if _this.MarsClient == nil {
		_this.MarsClient = MarsClient.Create()
	}

	// 停止時結束重試，避免清理期間又啟動新的雲端連線。
	for !_this.isStopping() {
		if _this.MarsClient.LoginWithProj(_url, _account, _pass, _proj) {
			Tools.Log.Print(Tools.LL_Info, "Init MarsCloud Client: true")
			return
		}
		Tools.Log.Print(Tools.LL_Info, "MarsCloud connect fail, Retry: %s", _url)
		_timer := time.NewTimer(5 * time.Second)
		select {
		case <-_timer.C:
		case <-_this.stopChan:
			_timer.Stop()
			return
		}
	}
}

// -------------------------------------------------------------------------------------
// doRegistry 執行服務註冊
func (_this *MarsService) doRegistry(_resetKey bool) bool {
	if _this.MarsClient != nil {
		_info := _this.ServiceInfo
		// 加入 PID 資訊
		_info.Put("pid", os.Getpid())

		if _this.MarsClient.RegistryService(_info.ToString(), _resetKey) {
			// 首次註冊成功才上傳 properties 並記錄一次 success log，避免每次 heartbeat 都重複
			if !_this.registered {
				// 與 MergePropertyFrom 互斥，避免 Property 序列化時被並行 mutation
				_this.propertyMu.RLock()
				_propStr := _this.Property.ToString()
				_this.propertyMu.RUnlock()
				_this.MarsClient.RegistryServiceProperties(_this.ServiceID, _propStr)
				_this.registered = true
				Tools.Log.Print(Tools.LL_Info, "Service registry success")
			}
			return true
		}
		Tools.Log.Print(Tools.LL_Info, "Service registry FAIL, try later ...")
	}
	return false
}

// -------------------------------------------------------------------------------------
// RegistryServerInfo 註冊服務資訊 (使用 JSONObject 實作)
func (_this *MarsService) RegistryServerInfo(_version string, _type string, _isOnline bool) {
	_this.ServiceType = strings.ToLower(_type)
	_this.ServiceVersion = _version

	// 填充 ServiceInfo (JSONObject)
	_this.ServiceInfo.Put("id", _this.ServiceID)
	_this.ServiceInfo.Put("name", _this.ServiceName)
	_this.ServiceInfo.Put("version", _this.ServiceVersion)
	_this.ServiceInfo.Put("type", "service."+_this.ServiceType)
	_this.ServiceInfo.Put("kernel_name", "[com.mars.cloudapp.go]")
	_this.ServiceInfo.Put("kernel_version", "0.2.13")
	_this.ServiceInfo.Put("vender", "MARS")
	_this.ServiceInfo.Put("timestamp", _this.SystemStartTime)
	_this.ServiceInfo.Put("web_hook", _this.webHook)
	_this.ServiceInfo.Put("owner", _this.account)
	_this.ServiceInfo.Put("ip", Tools.GetLocalIPv4Address())
	_this.ServiceInfo.Put("mac", Tools.GetLocalMACAddress(""))

	_this.ServiceInfo.Put("public", _this.Property.OptBoolean("is_public", true))
	_this.ServiceInfo.Put("initiative", true)
	_this.ServiceInfo.Put("is_online", _isOnline)

	Tools.Log.Print(Tools.LL_Debug, "Service Registered : %s", _version)

	// 定時同步 (Heartbeat)
	_this.syncTimer = time.NewTicker(20 * time.Second)
	go func() {
		for {
			select {
			case <-_this.syncTimer.C:
				_this.doRegistry(false)
			case <-_this.stopChan:
				return
			}
		}
	}()
}

// -------------------------------------------------------------------------------------
// MarsService MQTT 核心方法
// -------------------------------------------------------------------------------------

// InitMQTTClient 初始化 MQTT 連線設定
func (_this *MarsService) initMQTTClient(_url string) {

	if _url == "" || _this.MarsClient.AuthToken == "" {
		Tools.Log.Print(Tools.LL_Warning, "MQTT Init FAIL: URL or Token is empty")
		return
	}

	_tcpPort := _this.Property.OptInt("mqtt_tcp_port", 1883)
	_sslPort := _this.Property.OptInt("mqtt_ssl_port", 8883)
	_url = _this.resolveMQTTBrokerURL(_url, _tcpPort, _sslPort)
	Tools.Log.Print(Tools.LL_Info, "MQTT broker target: %s", _url)

	// 2. 設定專案 ID 與主題
	_topicID := _this.MarsClient.ProjID
	if _topicID == "" {
		_topicID = _this.MarsClient.Account
	}

	//_topic := _this.Property.OptString("mqtt_topic", _topicID+"/+/#")
	_topic := _this.Property.OptString("mqtt_topic", _topicID+"/event/+")
	if _topic == "" {
		Tools.Log.Print(Tools.LL_Debug, "MQTT is disabled: topic is empty")
	}

	_this.MQTT_Default_Topic = _topicID + "/event/" + _this.ServiceID
	_this.MQTT_AsyncTask_Topic = _topicID + "/service." + strings.ToLower(_this.ServiceType) + "/api"

	// 3. 建立連線選項
	_opts := MQTTClient.NewMQTTConnectOptions()
	_opts.SetServer(_url)
	_opts.SetCleanSession(true)
	_opts.SetAutomaticReconnect(true)
	_opts.SetUserName(_this.MarsClient.Account)
	_opts.SetClientID(_this.MarsClient.AuthToken)
	_opts.SetPassword([]byte(_this.MarsClient.AuthToken))
	_opts.SetKeepAliveInterval(30)
	_opts.SetConnectionTimeout(10)

	// 4. 建立 Client 並設定回調
	var _err error

	_mqttClient, _err := MQTTClient.Create()

	if _err != nil {
		Tools.Log.Print(Tools.LL_Error, "Create MQTT Client Fail : %s", _err.Error())
		return
	}

	_this.MQTTClient = _mqttClient
	_this.MQTTClient.SetCallback(&serviceCallback{service: _this})

	// 5. 執行連線
	if _err := _this.MQTTClient.Connect(_opts); _err != nil {
		Tools.Log.Print(Tools.LL_Error, "MQTT Connect Fail : %s", _err.Error())
		return
	}

	_this.ResetMQTTClient(_topic)
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) resolveMQTTBrokerURL(_rawURL string, _tcpPort int, _sslPort int) string {
	_parsed, _err := url.Parse(strings.TrimSpace(_rawURL))
	if _err != nil || _parsed == nil {
		return _rawURL
	}

	_host := _parsed.Hostname()
	if _host == "" {
		return _rawURL
	}

	switch _parsed.Scheme {
	case "https":
		return fmt.Sprintf("ssl://%s:%d", _host, _sslPort)
	case "http":
		return fmt.Sprintf("tcp://%s:%d", _host, _tcpPort)
	case "ssl", "tls", "mqtts":
		return fmt.Sprintf("ssl://%s:%d", _host, _sslPort)
	case "tcp", "mqtt":
		return fmt.Sprintf("tcp://%s:%d", _host, _tcpPort)
	case "ws", "wss":
		if _parsed.Port() != "" {
			return fmt.Sprintf("%s://%s", _parsed.Scheme, _parsed.Host)
		}
		if _parsed.Scheme == "wss" {
			return fmt.Sprintf("wss://%s:%d", _host, _this.Property.OptInt("mqtt_wss_port", 8884))
		}
		return fmt.Sprintf("ws://%s:%d", _host, _this.Property.OptInt("mqtt_ws_port", 1884))
	default:
		return _rawURL
	}
}

// -------------------------------------------------------------------------------------
// ResetMQTTClient 重新整理訂閱主題
func (_this *MarsService) ResetMQTTClient(_topic string) {
	if _this.MQTTClient == nil {
		return
	}

	// 在 Goroutine 中處理訂閱邏輯，避免阻塞主執行緒
	go func() {
		defer func() { recover() }()

		// 等待重連，最多 5 分鐘並聽 stopChan，避免服務關閉後 goroutine 永久殘留
		_ticker := time.NewTicker(1 * time.Second)
		defer _ticker.Stop()
		_timeout := time.After(5 * time.Minute)

		for !_this.MQTTClient.IsConnected() {
			select {
			case <-_ticker.C:
			case <-_this.stopChan:
				return
			case <-_timeout:
				Tools.Log.Print(Tools.LL_Warning, "MQTT reset wait timeout, abort")
				return
			}
		}

		Tools.Log.Print(Tools.LL_Info, "MQTT connection status : %v", _this.MQTTClient.IsConnected())

		if _this.MQTTClient.IsConnected() {
			_this.impl.OnMQTTConnected()

			// 訂閱必要主題
			_this.MQTTClient.Subscribe(_this.MQTT_Default_Topic, 0)
			_this.MQTTClient.Subscribe(_this.MQTT_AsyncTask_Topic, 0)

			if len(_topic) > 0 {

				if _this.MQTT_Default_Topic != _topic {
					_this.MQTTClient.Subscribe(_topic, 0)
				}

				_this.MQTT_Topic = _topic

				Tools.Log.Print(Tools.LL_Debug, "MQTT User Define Topic : %s", _this.MQTT_Topic)
			}
		}
	}()
}

//-------------------------------------------------------------------------------------
// 系統命令處理 (MQTT)
//-------------------------------------------------------------------------------------

// onDefaultMQTT 處理系統預設命令
func (_this *MarsService) onMQTTDefault(_topic, _payload string) {

	_msgObj := MarsJSON.NewJSONObject(_payload)
	_cmdObj := _msgObj
	_cmd := _msgObj.OptString("api", "")

	if _msgObj.Has("values") {
		_values := _msgObj.OptJSONArray("values")
		if _values != nil && _values.Length() > 0 {
			_cmdObj = _values.OptJSONObject(0)
			_cmd = _cmdObj.OptString("cmd", _cmdObj.OptString("api", ""))
		}
	}

	switch _cmd {
	case "reboot":
		_this.RestartService()
	case "shutdown":
		_this.ShutdownService()
	case "reset_properties":
		_this.ModifyProperties(_cmdObj.OptString("properties", ""))
	}
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) ModifyProperties(_payload string) {
	if _payload == "" {
		return
	}

	// 先寫到 .tmp 再 rename，避免中途失敗導致 properties 損毀後沒有可用設定
	_tmp := _this.PropertyFileName + ".tmp"
	if _err := os.WriteFile(_tmp, []byte(_payload), 0644); _err != nil {
		Tools.Log.Print(Tools.LL_Error, "Properties write fail: %s", _err.Error())
		return
	}
	if _err := os.Rename(_tmp, _this.PropertyFileName); _err != nil {
		Tools.Log.Print(Tools.LL_Error, "Properties rename fail: %s", _err.Error())
		os.Remove(_tmp)
		return
	}

	Tools.Log.Print(Tools.LL_Info, "Properties updated, restarting...")
	_this.RestartService()
}

//-------------------------------------------------------------------------------------
// HTTP 服務管理
//-------------------------------------------------------------------------------------

// ResetWebService 初始化或更新 Web 服務設定
func (_this *MarsService) ResetWebService() {

	_method := _this.Property.OptString("web_thread_method", "fixed")
	_coreSize := _this.Property.OptInt("web_core_count", 0)
	_maxSize := _this.Property.OptInt("web_thread_count", 500)
	_timeout := _this.Property.OptInt("web_thread_timeout", 500)

	if _this.HttpService != nil {
		// 如果服務已存在，更新其 Executor 設定
		_this.HttpService.InitExecutor(true, _method, _coreSize, _maxSize, _timeout)

	} else {
		// 建立新的 HttpService 實例
		_this.HttpService = HttpService.Create(
			_this.defaultHttpPort,
			_this.defaultHttpsPort,
			_this.ssl_Cert_File,
			_this.ssl_Key_File,
			_this.ssl_Key_Password,
		)
		// 初始化執行緒池邏輯 (在 Go 中主要為設定參數)
		_this.HttpService.InitExecutor(false, _method, _coreSize, _maxSize, _timeout)
	}

	_this.AddRestfulAPI("/system", HttpService.CreateHttpAPI_System(_this))

	Tools.Log.Print(Tools.LL_Debug, "Reset Web Service : %d/%d", _this.defaultHttpPort, _this.defaultHttpsPort)
	//Tools.Log.Print(Tools.LL_Info, "Set Web Thread : %s -> %d/%d", _method, _coreSize, _maxSize)
}

// -------------------------------------------------------------------------------------
// AddRestfulAPI 動態增加 API 路由
func (_this *MarsService) AddRestfulAPI(_uri string, _callback HttpService.HttpAPI_Callback) {

	if _uri != "" && _callback != nil {
		if _this.HttpService != nil {
			_this.HttpService.AddRestfulAPI(_uri, _callback)
		}
	}
}

// -------------------------------------------------------------------------------------
// RemoveRestfulAPI 移除指定的 API 路由
func (_this *MarsService) RemoveRestfulAPI(_uri string) {
	if _uri != "" {
		if _this.HttpService != nil {
			_this.HttpService.RemoveRestfulAPI(_uri)
		}
	}
}

//-------------------------------------------------------------------------------------
// Web 輔助工具
//-------------------------------------------------------------------------------------

// GetWebHook 獲取當前 Web 服務的 Hook 地址
func (_this *MarsService) getWebHook() string {

	_ip := Tools.GetLocalIPv4Address()
	_hook := _this.Property.OptString("url_hook", "")
	_hook = _this.Property.OptString("web_hook", _hook)

	if _hook != "" {
		return strings.Replace(strings.Replace(_hook, "127.0.0.1", _ip, 1), "localhost", _ip, 1)
	}

	return "http://" + _ip + fmt.Sprintf("%v", Tools.If(_this.defaultHttpPort == 80, "", fmt.Sprintf(":%v", _this.defaultHttpPort)))
}

// ------------------------------------------------------------------------------------
func (_this *MarsService) GetProperty() *MarsJSON.JSONObject {
	if _this.Property != nil {
		return _this.Property
	}

	return MarsJSON.NewJSONObject("{}")
}

// ------------------------------------------------------------------------------------
func (_this *MarsService) MergePropertyFrom(_ext *MarsJSON.JSONObject) {
	_this.propertyMu.Lock()
	defer _this.propertyMu.Unlock()
	if _this.Property != nil {
		_this.Property.MergeFrom(_ext)
	}
}

// ------------------------------------------------------------------------------------
func (_this *MarsService) OnUpdateProperty() {

	if _this.impl != nil {

		_this.impl.OnPropertyChange(_this.Property)
	}
}

// ------------------------------------------------------------------------------------
// SendResponse 靜態工具的服務層包裝
func (_this *MarsService) SendResponse(_w http.ResponseWriter, _no int, _contentType string, _content []byte) {
	HttpService.SendResponse(_w, _no, _contentType, _content)
}

//-------------------------------------------------------------------------------------
// 資源管理 (GC / WebService)
//-------------------------------------------------------------------------------------

// ResetAutoGC 設定自動記憶體回收
func (_this *MarsService) ResetAutoGC() {
	_gcInterval := _this.Property.OptInt("auto_gc", 1800) // 預設 30 分鐘
	if _gcInterval > 0 {
		_this.defaultTimer = time.NewTicker(time.Duration(_gcInterval) * time.Second)
		go func() {
			for {
				select {
				case <-_this.defaultTimer.C:
					runtime.GC()
					Tools.Log.Print(Tools.LL_Debug, "System GC executed")
				case <-_this.stopChan:
					return
				}
			}
		}()
	}
}

//-------------------------------------------------------------------------------------
// 關閉邏輯
//-------------------------------------------------------------------------------------

func (_this *MarsService) isStopping() bool {
	select {
	case <-_this.stopChan:
		return true
	default:
		return false
	}
}

func (_this *MarsService) StopService() bool {
	if !_this.lifecycleMu.TryLock() {
		return false
	}
	defer _this.lifecycleMu.Unlock()
	_this.stopService()
	return true
}

// 呼叫端持有 lifecycleMu；先取消背景工作，清理 callback 與網路資源只執行一次。
func (_this *MarsService) stopService() {
	_this.stopOnce.Do(func() {
		close(_this.stopChan)
		if _this.syncTimer != nil {
			_this.syncTimer.Stop()
		}
		if _this.defaultTimer != nil {
			_this.defaultTimer.Stop()
		}
		defer _this.CloseNetService()
		if _this.impl != nil {
			_this.impl.BeforeServiceStop()
		}
		if _this.ServiceInfo != nil {
			_this.ServiceInfo.Put("is_online", false)
			_this.doRegistry(false)
		}
		Tools.Log.Print(Tools.LL_Info, "Service Stopped")
	})
}

// 程序即將退出時，限制自訂 callback／遠端離線通知的等待時間。
func (_this *MarsService) stopBeforeExit() {
	_done := make(chan struct{})
	go func() {
		defer close(_done)
		defer func() {
			if _r := recover(); _r != nil {
				Tools.Log.Print(Tools.LL_Error, "服務停止清理發生異常: %v", _r)
			}
		}()
		_this.stopService()
	}()
	_timer := time.NewTimer(30 * time.Second)
	defer _timer.Stop()
	select {
	case <-_done:
	case <-_timer.C:
		Tools.Log.Print(Tools.LL_Warning, "服務停止清理超過 30 秒，退出程序以釋放資源")
	}
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) CloseNetService() {
	// 先收 HTTP listener，避免重啟時新子程序撞到舊 process 仍在 bind 的 port
	if _this.HttpService != nil {
		_this.HttpService.Close()
	}
	if _this.MQTTClient != nil {
		_this.MQTTClient.Disconnect(250)
	}
	if _this.LocalMQTTServer != nil {
		if _err := _this.LocalMQTTServer.Close(); _err != nil {
			Tools.Log.Print(Tools.LL_Error, "Close local MQTT server fail: %s", _err.Error())
		}
		_this.LocalMQTTServer = nil
	}
	Tools.Log.Print(Tools.LL_Info, "Network services closed")
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) ShutdownService() {
	// 關機訊號不可遺失：若啟動或重啟交接正在進行，等待該操作完成後再清理。
	_this.lifecycleMu.Lock()
	defer _this.lifecycleMu.Unlock()
	_this.stopBeforeExit()
	os.Exit(0)
}

// -------------------------------------------------------------------------------------
// 重啟管理
// -------------------------------------------------------------------------------------
func (_this *MarsService) RestartService() {
	if !_this.lifecycleMu.TryLock() {
		Tools.Log.Print(Tools.LL_Warning, "服務正在執行啟停操作，略過重複重啟要求")
		return
	}
	defer _this.lifecycleMu.Unlock()
	if _this.isStopping() {
		return
	}

	Tools.Log.Print(Tools.LL_Warning, "Service is preparing to restart...")
	// 必須先嘗試交接；失敗時不停止 listener、背景工作或自訂業務資源。
	// Unix 以原 PID exec，成功後不返回；Windows 新程序會等待目前程序退出。
	if _err := Tools.RestartItSelf(); _err != nil {
		Tools.Log.Print(Tools.LL_Error, "重啟失敗，目前服務繼續運作: %v", _err)
		return
	}
	_this.stopBeforeExit()
	os.Exit(0)
}

// -------------------------------------------------------------------------------------
// ResetAutoRestart 初始化自動定時重啟任務
func (_this *MarsService) ResetAutoRestart() {

	// 如果沒有有效的重啟時間，則直接返回
	if len(_this.autoRestartMinutes) == 0 {
		return
	}

	Tools.Log.Print(Tools.LL_Info, "Auto restart at : %v", _this.autoRestartMinutes)

	// 開啟一個 Goroutine 定時檢查時間
	go func() {
		// 任何 panic 都不能讓監控悄悄結束，否則自動重啟會永久失效
		defer func() {
			if _r := recover(); _r != nil {
				Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("Auto-restart monitor panic: %v", _r))
			}
		}()

		_ticker := time.NewTicker(1 * time.Second)
		defer _ticker.Stop()

		// 以分鐘為比對精度，避免 ticker 因 GC/排程延遲跳過該秒整天 miss
		// 同分鐘已觸發過則跳過，防止 RestartService 失敗時每秒重複觸發
		_lastFireMinute := ""

		for {
			select {
			case <-_ticker.C:
				_uptime := time.Now().UnixMilli() - _this.SystemStartTime
				if _uptime < 60000 {
					continue
				}

				_now := time.Now().In(_this.restartLocation)
				_currentMinute := _now.Format("15:04")
				_fireMinute := _now.Format("2006-01-02 15:04")
				if _fireMinute == _lastFireMinute {
					continue
				}

				for _, _target := range _this.autoRestartMinutes {
					if _currentMinute == _target {
						Tools.Log.Print(Tools.LL_Warning, "Restart time reached: "+_target)
						_lastFireMinute = _fireMinute
						_this.RestartService()
						// 成功時程序會 os.Exit(0)；失敗時繼續監控等下個排程
						break
					}
				}
			case <-_this.stopChan:
				return
			}
		}
	}()
}

//-------------------------------------------------------------------------------------
// 功能包裝 (Wrappers)
//-------------------------------------------------------------------------------------

func (_this *MarsService) PutData(_uuid, _suid string, _data *MarsJSON.JSONObject) bool {
	if _this.MarsClient != nil {
		return _this.MarsClient.PutData(_uuid, _suid, _data)
	}
	return false
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) PushLineMessage(_target, _thisg string) bool {
	if _this.MarsClient != nil {
		//return _this.MarsClient.PushLineMessage(_target, _thisg)
	}
	return false
}

//-------------------------------------------------------------------------------------
