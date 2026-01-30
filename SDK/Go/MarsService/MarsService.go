package MarsService

//-------------------------------------------------------------------------------------
import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/AsyncTaskProcessor"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/HttpService"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MQTTClient"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsClient"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
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
func (_this *serviceCallback) OnMessageArrived(_topic string, _thisg *MQTTClient.MQTTMessage) {
	defer func() {
		if _r := recover(); _r != nil {
			Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("MQTT Message Error: %v", _r))
		}
	}()

	_payload := string(_thisg.GetPayload())

	// 1. 處理預設系統主題
	if _this.service.MQTT_Default_Topic == _topic {
		_this.service.OnDefaultMQTTMessage(_topic, _payload)
		if _this.service.MQTT_Default_Topic == _this.service.MQTT_Topic {
			_this.service.impl.OnMQTTMessage(_topic, _payload)
		}
	} else {
		// 2. 處理非同步任務或是具體業務主題
		if _this.service.AsyncTaskProcessor != nil && strings.Contains(_topic, "/service."+strings.ToLower(_this.service.ServiceType)) {
			_this.service.AsyncTaskProcessor.OnMQTTMessage(_topic, _payload)
		} else {
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
	ssl_Key_File     string
	ssl_Key_Password string

	SystemStartTime      int64
	IsDebugData          bool
	RestartAfterConflict bool

	syncTimer       *time.Ticker
	defaultTimer    *time.Ticker // 用於 AutoGC
	serviceRegCheck *time.Timer  // 檢查註冊狀態

	stopChan chan struct{}

	autoRestartTime MarsJSON.JSONArray
	webHook         string
	impl            IMarsService // 指向具體的實作物件_
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
	_this.ssl_Key_File = _this.Property.OptString("ssl_key", "")
	_this.ssl_Key_Password = _this.Property.OptString("ssl_key_password", "")

	_this.ServiceName = _this.Property.OptString("service_name", "Unknown Service")
	_this.ServiceID = fmt.Sprintf("%s-%d", Tools.GetMachineID(), _this.defaultHttpPort)

	_this.webHook = _this.Property.OptString("url_hook", "")
	_this.autoRestartTime = *_this.Property.OptJSONArray("restart_time") //["6:00:00", "14:12:24"]

	_url := _this.Property.OptString("mars_cloud_url", "")
	_account := _this.Property.OptString("mars_cloud_account", "")
	_pwd := _this.Property.OptString("mars_cloud_password", "")
	_proj := _this.Property.OptString("mars_cloud_proj", "")

	fmt.Printf("\n------------------------------------\n")
	fmt.Printf("\n %s \n", _this.ServiceName)
	fmt.Printf("\n------------------------------------\n\n")

	Tools.Log.Print(Tools.LL_Info, "Service ID : %s", _this.ServiceID)
	Tools.Log.Print(Tools.LL_Info, "Process ID : %d", Tools.GetPID(nil))

	_this.initMarsClient(_url, _account, _pwd, _proj)
	_this.ResetWebService()
}

// -------------------------------------------------------------------------------------
// killProcessByPort 根據埠號找出並關閉進程 (支援 Windows, Mac, Linux)
// -------------------------------------------------------------------------------------
func (_this *MarsService) killProcessByPort(_port int) {
	var _pid string

	if Tools.IsMSWindow() {
		// Windows: 透過 netstat 找出佔用該 port 且處於 LISTENING 狀態的 PID
		// 指令範例: netstat -ano | findstr :80 | findstr LISTENING
		_cmd := fmt.Sprintf("netstat -ano | findstr :%d | findstr LISTENING", _port)
		_out := Tools.ShellCMDSync(_cmd)
		_lines := strings.Split(strings.TrimSpace(_out), "\n")

		if len(_lines) > 0 && _lines[0] != "" {
			// Windows netstat 輸出格式最後一欄位通常是 PID
			_parts := strings.Fields(_lines[0])
			if len(_parts) > 0 {
				_pid = _parts[len(_parts)-1]
			}
		}

	} else {
		// Linux & macOS: 使用 lsof 直接取得 PID (-t 代表僅輸出 PID)
		_cmd := fmt.Sprintf("lsof -t -i:%d", _port)
		_pid = strings.TrimSpace(Tools.ShellCMDSync(_cmd))
	}

	// 如果有找到 PID，則執行殺掉進程的動作
	if _pid != "" && _pid != "0" {
		Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Found PID %s occupying port %d. Killing it...", _pid, _port))
		Tools.KillProcess(_pid)
	}
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) checPortConflict() {

	// 檢查端口衝突並嘗試自動排除
	if Tools.IsPortInUsing(_this.defaultHttpPort) {
		// 執行強制清理
		_this.killProcessByPort(_this.defaultHttpPort)

		// 給予作業系統短暫的時間釋放 Socket
		time.Sleep(3 * time.Second)

		// 再次檢查，如果還是被佔用，則執行原有的衝突處理邏輯
		if Tools.IsPortInUsing(_this.defaultHttpPort) {
			Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("Unable to clear Port %d, check permissions.", _this.defaultHttpPort))
			if _this.Property.OptBoolean("conflict_restart", false) {
				_this.RestartService()
				return
			} else {
				os.Exit(0)
			}
		}
	}
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) Start() {

	// 檢查端口衝突
	_this.checPortConflict()

	if _this.HttpService != nil {

		_this.HttpService.SetRootPath(_this.Property.OptString("web_path", "./website"))
		_this.HttpService.SetDefaultCacheControl("public, max-age=43200")
		_this.HttpService.Run()
	}

	_this.ResetAutoRestart()
	_this.ResetAutoGC()

	Tools.Log.Print(Tools.LL_Info, "Service Start : "+_this.ServiceName)
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
		Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Get Closing Singal : %v, clean up ...", _sig))

		_this.impl.BeforeServiceStop()
		_this.ServiceInfo.Put("is_online", false)
		_this.DoRegistry(false)

		Tools.Log.Print(Tools.LL_Info, "Clean up finish, process exit")
		Tools.Log.Print(Tools.LL_Info, "- ")

		os.Exit(0)
	}()
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

	// 嘗試登入，失敗則持續重試 (模擬 Java while 邏輯)
	for !_this.MarsClient.LoginWithProj(_url, _account, _pass, _proj) {
		Tools.Log.Print(Tools.LL_Info, "MarsCloud connect fail, Retry: %s", _url)
		time.Sleep(5 * time.Second)
	}

	Tools.Log.Print(Tools.LL_Info, "Init MarsCloud Client: true")

	// 初始化 MQTT，使用 MetaClient 取得的 Server URL
	_this.initMQTTClient(_this.MarsClient.GetServerURL())
}

// -------------------------------------------------------------------------------------
// DoRegistry 執行服務註冊
func (_this *MarsService) DoRegistry(_resetKey bool) bool {
	if _this.MarsClient != nil {
		_info := _this.ServiceInfo
		// 加入 PID 資訊
		_info.Put("pid", os.Getpid())

		if _this.MarsClient.RegistryService(_info.ToString(), _resetKey) {
			if _this.serviceRegCheck != nil {
				_this.serviceRegCheck.Stop()
				_this.serviceRegCheck = nil
				// 註冊屬性配置
				_this.MarsClient.RegistryServiceProperties(_this.ServiceID, _this.Property.ToString())

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
	_this.ServiceInfo.Put("kernel_version", "0.1.0")
	_this.ServiceInfo.Put("vender", "MARS")
	_this.ServiceInfo.Put("timestamp", _this.SystemStartTime)
	_this.ServiceInfo.Put("web_hook", _this.webHook)
	_this.ServiceInfo.Put("owner", _this.MarsClient.Account)
	_this.ServiceInfo.Put("ip", Tools.GetLocalIPv4Address())
	_this.ServiceInfo.Put("mac", Tools.GetLocalMACAddress(""))

	_this.ServiceInfo.Put("public", false)
	_this.ServiceInfo.Put("initiative", true)
	_this.ServiceInfo.Put("is_online", _isOnline)

	Tools.Log.Print(Tools.LL_Info, "Service Registered: %s", _version)

	// 定時同步 (Heartbeat)
	_this.syncTimer = time.NewTicker(20 * time.Second)
	go func() {
		for {
			select {
			case <-_this.syncTimer.C:
				_this.DoRegistry(false)
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

	// 1. 協定與埠號轉換 (http -> tcp:1883, https -> ssl:8883)
	if strings.Contains(_url, ":") && strings.LastIndex(_url, ":") > 5 {
		_url = _url[:strings.LastIndex(_url, ":")]
	}
	if strings.HasPrefix(_url, "http://") {
		_url = strings.Replace(_url, "http", "tcp", 1) + ":1883"
	} else if strings.HasPrefix(_url, "https://") {
		_url = strings.Replace(_url, "https", "ssl", 1) + ":8883"
	}

	// 2. 設定專案 ID 與主題
	_topicID := _this.MarsClient.ProjID
	if _topicID == "" {
		_topicID = _this.MarsClient.Account
	}

	_topic := _this.Property.OptString("mqtt_topic", _topicID+"/+/#")
	if _topic == "" {
		Tools.Log.Print(Tools.LL_Warning, "MQTT is disabled: topic is empty")
		return
	}

	_this.MQTT_Default_Topic = _topicID + "/event/" + _this.ServiceID
	_this.MQTT_AsyncTask_Topic = _topicID + "/service." + strings.ToLower(_this.ServiceType) + "/+"

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
// ResetMQTTClient 重新整理訂閱主題
func (_this *MarsService) ResetMQTTClient(_topic string) {
	if _this.MQTTClient == nil {
		return
	}

	// 在 Goroutine 中處理訂閱邏輯，避免阻塞主執行緒
	go func() {
		defer func() { recover() }()

		if _topic == "" {
			_this.MQTTClient.Disconnect(250)
			_this.MQTT_Topic = ""
			return
		}

		// 模擬 Java 版的重連與主題重整邏輯
		_tick := 0
		for {
			if _this.MQTTClient.IsConnected() {
				break
			}
			time.Sleep(1 * time.Second)
			_tick++
			if _tick >= 5 {
				_tick = 0
				// 此處依賴底層自動重連，或手動觸發連線
			}
		}

		if _this.MQTTClient.IsConnected() {
			_this.impl.OnMQTTConnected()

			// 訂閱必要主題
			_this.MQTTClient.Subscribe(_this.MQTT_Default_Topic, 0)
			_this.MQTTClient.Subscribe(_this.MQTT_AsyncTask_Topic, 0)

			if _this.MQTT_Default_Topic != _topic {
				_this.MQTTClient.Subscribe(_topic, 0)
			}

			_this.MQTT_Topic = _topic
			Tools.Log.Print(Tools.LL_Debug, "MQTT current Topic : %s", _this.MQTT_Topic)
		}

		Tools.Log.Print(Tools.LL_Info, "MQTT connection status : %v", _this.MQTTClient.IsConnected())
	}()
}

//-------------------------------------------------------------------------------------
// 系統命令處理 (MQTT)
//-------------------------------------------------------------------------------------

// OnDefaultMQTTMessage 處理系統預設命令
func (_this *MarsService) OnDefaultMQTTMessage(_topic, _payload string) {
	_thisgObj := MarsJSON.NewJSONObject(_payload)
	// 取得 values 陣列中的第一個指令
	_values := _thisgObj.OptJSONArray("values")
	if _values != nil && _values.Length() > 0 {
		_cmdObj := _values.OptJSONObject(0)
		_cmd := _cmdObj.OptString("cmd", "")

		switch _cmd {
		case "reboot":
			_this.RestartService()
		case "shutdown":
			_this.ShutdownService()
		case "reset_properties":
			_this.ModifyProperties(_cmdObj.OptString("properties", ""))
		}
	}
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) ModifyProperties(_payload string) {
	if _payload != "" {
		// 儲存新的設定檔
		os.WriteFile(_this.PropertyFileName, []byte(_payload), 0644)
		Tools.Log.Print(Tools.LL_Info, "Properties updated, restarting...")
		_this.RestartService()
	}
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
func (_this *MarsService) GetWebHook() string {

	_ip := _this.ServiceInfo.OptString("ip", "")

	if _this.webHook != "" {
		return strings.Replace(strings.Replace(_this.webHook, "127.0.0.0", _ip, 1), "localhost", _ip, 1)
	}

	return fmt.Sprintf("http://%s:%d", _ip, _this.defaultHttpPort)
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
// SendRespone 靜態工具的服務層包裝
func (_this *MarsService) SendRespone(_w http.ResponseWriter, _no int, _contentType string, _content []byte) {
	HttpService.SendRespone(_w, _no, _contentType, _content)
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
			for range _this.defaultTimer.C {
				runtime.GC()
				Tools.Log.Print(Tools.LL_Debug, "System GC executed")
			}
		}()
	}
}

// -------------------------------------------------------------------------------------
// 重啟管理
// -------------------------------------------------------------------------------------
func (_this *MarsService) RestartService() {

	Tools.Log.Print(Tools.LL_Warning, "Service is preparing to restart...")

	// 1. 執行停止前的清理邏輯
	if _this.impl != nil {
		_this.impl.BeforeServiceStop()
	}

	// 2. 通知雲端服務目前為離線狀態並執行最後一次註冊同步
	if _this.ServiceInfo != nil {
		_this.ServiceInfo.Put("is_online", false)
		_this.DoRegistry(false)
	}

	// 3. 關閉網路連線資源
	_this.CloseNetService()

	// 4. 呼叫 Tools 中的實體重啟邏輯
	_err := Tools.RestartItSelf()

	if _err != nil {
		Tools.Log.Print(Tools.LL_Error, "Restart failed: %s", _err.Error())
		return
	}

	// 5. 成功啟動新程序後，退出當前程序
	os.Exit(0)
}

// -------------------------------------------------------------------------------------
// ResetAutoRestart 初始化自動定時重啟任務
func (_this *MarsService) ResetAutoRestart() {

	// 如果沒有設定重啟時間，則直接返回
	if _this.autoRestartTime.Length() <= 0 {
		return
	}

	Tools.Log.Print(Tools.LL_Info, "Auto restart at : %v", _this.autoRestartTime.ToString())

	// 開啟一個 Goroutine 定時檢查時間
	go func() {
		_ticker := time.NewTicker(1 * time.Second)
		defer _ticker.Stop()

		for {
			select {
			case <-_ticker.C:
				_uptime := time.Now().UnixMilli() - _this.SystemStartTime

				// 若運作未滿 60 秒 (60,000 ms)，跳過此次檢查，避免重複重啟循環
				if _uptime < 60000 {
					continue
				}

				// 獲取當前時間字串 (格式如 "14:30:00")
				_currentTime := time.Now().Format("15:04:05")

				// 遍歷 JSONArray 檢查是否有匹配的時間點
				for _i := 0; _i < _this.autoRestartTime.Length(); _i++ {
					_targetTime := _this.autoRestartTime.OptString(_i, "")

					if _currentTime == _targetTime {
						Tools.Log.Print(Tools.LL_Warning, "Restart time reached: "+_targetTime)
						_this.RestartService()
						return // 觸發重啟後結束此監控
					}
				}
			case <-_this.stopChan:
				// 接收到停止訊號則結束監控
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
// 關閉邏輯
//-------------------------------------------------------------------------------------

func (_this *MarsService) StopService() bool {
	_this.CloseNetService()
	Tools.Log.Print(Tools.LL_Info, "Service Stoped")
	return true
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) ShutdownService() {
	_this.StopService()
	os.Exit(0)
}

// -------------------------------------------------------------------------------------
func (_this *MarsService) CloseNetService() {
	if _this.MQTTClient != nil {
		_this.MQTTClient.Disconnect(250)
	}
	Tools.Log.Print(Tools.LL_Info, "Network services closed")
}

//-------------------------------------------------------------------------------------
