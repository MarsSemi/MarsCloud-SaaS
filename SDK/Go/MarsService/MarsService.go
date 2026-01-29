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
	_ms *MarsService
}

// -------------------------------------------------------------------------------------
func (_cb *serviceCallback) OnConnected() {
	Tools.Log.Print(Tools.LL_Info, "MQTT connected")
	_cb._ms.impl.OnMQTTConnected()
}

// -------------------------------------------------------------------------------------
func (_cb *serviceCallback) OnConnectionLost(_err error) {
	Tools.Log.Print(Tools.LL_Warning, "MQTT connection lost")
	_cb._ms.impl.OnMQTTConnectionLost(_err)
	_cb._ms.ResetMQTTClient(_cb._ms.MQTT_Topic)
}

// -------------------------------------------------------------------------------------
func (_cb *serviceCallback) OnDeliveryComplete(_token string) {}

// -------------------------------------------------------------------------------------
func (_cb *serviceCallback) OnMessageArrived(_topic string, _msg *MQTTClient.MQTTMessage) {
	defer func() {
		if _r := recover(); _r != nil {
			Tools.Log.Print(Tools.LL_Error, fmt.Sprintf("MQTT Message Error: %v", _r))
		}
	}()

	_payload := string(_msg.GetPayload())

	// 1. 處理預設系統主題
	if _cb._ms.MQTT_Default_Topic == _topic {
		_cb._ms.OnDefaultMQTTMessage(_topic, _payload)
		if _cb._ms.MQTT_Default_Topic == _cb._ms.MQTT_Topic {
			_cb._ms.impl.OnMQTTMessage(_topic, _payload)
		}
	} else {
		// 2. 處理非同步任務或是具體業務主題
		if _cb._ms.AsyncTaskProcessor != nil && strings.Contains(_topic, "/service."+strings.ToLower(_cb._ms.ServiceType)) {
			_cb._ms.AsyncTaskProcessor.OnMQTTMessage(_topic, _payload)
		} else {
			_cb._ms.impl.OnMQTTMessage(_topic, _payload)
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
	SyncTimer            *time.Ticker
	DefaultTimer         *time.Ticker // 用於 AutoGC
	ServiceRegCheck      *time.Timer  // 檢查註冊狀態

	stopChan chan struct{}

	webHook string
	impl    IMarsService // 指向具體的實作物件_
}

// -------------------------------------------------------------------------------------
func Create(_propertyFileName string, _impl IMarsService) *MarsService {
	_ms := &MarsService{
		PropertyFileName: _propertyFileName,
		SystemStartTime:  time.Now().UnixMilli(),
		ServiceInfo:      MarsJSON.NewJSONObject(""),
		stopChan:         make(chan struct{}),
		impl:             _impl,
	}

	_ms.initCloseHook()
	_ms.init(_propertyFileName)

	return _ms
}

//-------------------------------------------------------------------------------------
// 核心初始化邏輯
//-------------------------------------------------------------------------------------

func (_ms *MarsService) init(_propertyFileName string) {

	_ms.Property = MarsJSON.NewJSONObject(Tools.File2String(_propertyFileName))

	// 初始化基本變數
	_ms.defaultHttpPort = _ms.Property.OptInt("http_port", 80)
	_ms.defaultHttpsPort = _ms.Property.OptInt("https_port", 433)
	_ms.ssl_Key_File = _ms.Property.OptString("ssl_key", "")
	_ms.ssl_Key_Password = _ms.Property.OptString("ssl_key_password", "")

	_ms.ServiceName = _ms.Property.OptString("service_name", "Unknown Service")
	_ms.ServiceID = fmt.Sprintf("%s-%d", Tools.GetMachineID(), _ms.defaultHttpPort)

	_ms.webHook = _ms.Property.OptString("url_hook", "")

	_url := _ms.Property.OptString("mars_cloud_url", "")
	_account := _ms.Property.OptString("mars_cloud_account", "")
	_pwd := _ms.Property.OptString("mars_cloud_password", "")
	_proj := _ms.Property.OptString("mars_cloud_proj", "")

	fmt.Printf("\n------------------------------------\n")
	fmt.Printf("\n %s \n", _ms.ServiceName)
	fmt.Printf("\n------------------------------------\n\n")

	_ms.initMarsClient(_url, _account, _pwd, _proj)
	_ms.ResetWebService()
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) Start() {

	// 檢查端口衝突
	if Tools.IsPortInUsing(_ms.defaultHttpPort) {
		Tools.Log.Print(Tools.LL_Warning, fmt.Sprintf("Port %d is in using", _ms.defaultHttpPort))
		if _ms.Property.OptBoolean("conflict_restart", true) {
			_ms.RestartService()
		} else {
			os.Exit(0)
		}
	}

	if _ms.HttpService != nil {

		_ms.HttpService.SetRootPath(_ms.Property.OptString("web_path", "./website"))
		_ms.HttpService.SetDefaultCacheControl("public, max-age=43200")
		_ms.HttpService.Run()
	}

	_ms.ResetAutoRestart()
	_ms.ResetAutoGC()

	Tools.Log.Print(Tools.LL_Info, "Service Start : "+_ms.ServiceName)
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) initCloseHook() {
	// 建立一個頻道來接收作業系統訊號
	_sigChan := make(chan os.Signal, 1)

	// 註冊要監聽的訊號 (SIGINT = Ctrl+C, SIGTERM = 停止訊號)
	signal.Notify(_sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 開啟一個 Goroutine 在背景等待訊號
	go func() {
		_sig := <-_sigChan // 這會阻塞直到收到訊號

		Tools.Log.Print(Tools.LL_Info, "- ")
		Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Get Closing Singal : %v, clean up ...", _sig))

		_ms.impl.BeforeServiceStop()
		_ms.ServiceInfo.Put("is_online", false)
		_ms.DoRegistry(false)

		Tools.Log.Print(Tools.LL_Info, "Clean up finish, process exit")
		Tools.Log.Print(Tools.LL_Info, "- ")

		os.Exit(0)
	}()
}

// -------------------------------------------------------------------------------------
// InitMarsClient 初始化與 MarsCloud 的連線
func (_ms *MarsService) initMarsClient(_url, _account, _pass, _proj string) {
	if _url == "" || _account == "" || _pass == "" {
		return
	}

	if _ms.MarsClient == nil {
		_ms.MarsClient = MarsClient.Create()
	}

	// 嘗試登入，失敗則持續重試 (模擬 Java while 邏輯)
	for !_ms.MarsClient.LoginWithProj(_url, _account, _pass, _proj) {
		Tools.Log.Print(Tools.LL_Info, "MarsCloud connect fail, Retry: "+_url)
		time.Sleep(5 * time.Second)
	}

	Tools.Log.Print(Tools.LL_Info, "Init MarsCloud Client: true")

	// 初始化 MQTT，使用 MetaClient 取得的 Server URL
	_ms.initMQTTClient(_ms.MarsClient.GetServerURL())
}

// -------------------------------------------------------------------------------------
// DoRegistry 執行服務註冊
func (_ms *MarsService) DoRegistry(_resetKey bool) bool {
	if _ms.MarsClient != nil {
		_info := _ms.ServiceInfo
		// 加入 PID 資訊
		_info.Put("pid", os.Getpid())

		if _ms.MarsClient.RegistryService(_info.ToString(), _resetKey) {
			if _ms.ServiceRegCheck != nil {
				_ms.ServiceRegCheck.Stop()
				_ms.ServiceRegCheck = nil
				// 註冊屬性配置
				_ms.MarsClient.RegistryServiceProperties(_ms.ServiceID, _ms.Property.ToString())

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
func (_ms *MarsService) RegistryServerInfo(_version string, _type string, _isOnline bool) {
	_ms.ServiceType = strings.ToLower(_type)
	_ms.ServiceVersion = _version

	// 填充 ServiceInfo (JSONObject)
	_ms.ServiceInfo.Put("id", _ms.ServiceID)
	_ms.ServiceInfo.Put("name", _ms.ServiceName)
	_ms.ServiceInfo.Put("version", _ms.ServiceVersion)
	_ms.ServiceInfo.Put("type", "service."+_ms.ServiceType)
	_ms.ServiceInfo.Put("kernel_name", "[com.mars.cloudapp.go]")
	_ms.ServiceInfo.Put("kernel_version", "0.1.0")
	_ms.ServiceInfo.Put("vender", "MARS")
	_ms.ServiceInfo.Put("timestamp", _ms.SystemStartTime)
	_ms.ServiceInfo.Put("web_hook", _ms.webHook)
	_ms.ServiceInfo.Put("owner", _ms.MarsClient.Account)
	_ms.ServiceInfo.Put("ip", Tools.GetLocalIPv4Address())
	_ms.ServiceInfo.Put("mac", Tools.GetLocalMACAddress(""))

	_ms.ServiceInfo.Put("public", false)
	_ms.ServiceInfo.Put("initiative", true)
	_ms.ServiceInfo.Put("is_online", _isOnline)

	/*
		_Source.put("id", _ServiceID);
		_Source.put("name", _ServiceName);
		_Source.put("description", _ServiceDescription);
		_Source.put("owner", _Owner);
		_Source.put("vender", _Vender);
		_Source.put("version", _Version);
		_Source.put("kernel_version", _Kernel_Version);
		_Source.put("kernel_name", _Kernel_Name);
		_Source.put("start_time", _StartRunningTime);
		_Source.put("mac", _MACAddress);
		_Source.put("white_list", _whiteList);
		_Source.put("host", _Host);
		_Source.put("type", _Type);
		_Source.put("web_hook", _WebHook);
		_Source.put("directory", _Directory.replaceAll("\\\\", "/"));
		_Source.put("vid", _VID);
		_Source.put("timestamp", _StartRunningTimeStamp);
		_Source.put("city", _City);
		_Source.put("auth_due_date", _AuthDueDate);
		_Source.put("auth_target", _AuthTarget);
		_Source.put("lat", _Lat);
		_Source.put("lng", _Lng);
		_Source.put("enable_white_list", _IsEnableWhiteList);
		_Source.put("public", _IsPublic);
		_Source.put("online", _IsOnline);
		_Source.put("initiative", _IsInitiative);
	*/

	Tools.Log.Print(Tools.LL_Info, "Service Registered: "+_version)

	// 定時同步 (Heartbeat)
	_ms.SyncTimer = time.NewTicker(20 * time.Second)
	go func() {
		for {
			select {
			case <-_ms.SyncTimer.C:
				_ms.DoRegistry(false)
			case <-_ms.stopChan:
				return
			}
		}
	}()
}

// -------------------------------------------------------------------------------------
// MarsService MQTT 核心方法
// -------------------------------------------------------------------------------------
// InitMQTTClient 初始化 MQTT 連線設定
func (_ms *MarsService) initMQTTClient(_url string) {

	if _url == "" || _ms.MarsClient.AuthToken == "" {
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
	_topicID := _ms.MarsClient.ProjID
	if _topicID == "" {
		_topicID = _ms.MarsClient.Account
	}

	_topic := _ms.Property.OptString("mqtt_topic", _topicID+"/+/#")
	if _topic == "" {
		Tools.Log.Print(Tools.LL_Warning, "MQTT is disabled: topic is empty")
		return
	}

	_ms.MQTT_Default_Topic = _topicID + "/event/" + _ms.ServiceID
	_ms.MQTT_AsyncTask_Topic = _topicID + "/service." + strings.ToLower(_ms.ServiceType) + "/+"

	// 3. 建立連線選項
	_opts := MQTTClient.NewMQTTConnectOptions()
	_opts.SetServer(_url)
	_opts.SetCleanSession(true)
	_opts.SetAutomaticReconnect(true)
	_opts.SetUserName(_ms.MarsClient.Account)
	_opts.SetClientID(_ms.MarsClient.AuthToken)
	_opts.SetPassword([]byte(_ms.MarsClient.AuthToken))
	_opts.SetKeepAliveInterval(30)
	_opts.SetConnectionTimeout(10)

	// 4. 建立 Client 並設定回調
	var _err error

	_mqttClient, _err := MQTTClient.Create()

	if _err != nil {
		Tools.Log.Print(Tools.LL_Error, "Create MQTT Client Fail : "+_err.Error())
		return
	}

	_ms.MQTTClient = _mqttClient
	_ms.MQTTClient.SetCallback(&serviceCallback{_ms: _ms})

	// 5. 執行連線
	if _err := _ms.MQTTClient.Connect(_opts); _err != nil {
		Tools.Log.Print(Tools.LL_Error, "MQTT Connect Fail : "+_err.Error())
		return
	}

	_ms.ResetMQTTClient(_topic)
}

// -------------------------------------------------------------------------------------
// ResetMQTTClient 重新整理訂閱主題
func (_ms *MarsService) ResetMQTTClient(_topic string) {
	if _ms.MQTTClient == nil {
		return
	}

	// 在 Goroutine 中處理訂閱邏輯，避免阻塞主執行緒
	go func() {
		defer func() { recover() }()

		if _topic == "" {
			_ms.MQTTClient.Disconnect(250)
			_ms.MQTT_Topic = ""
			return
		}

		// 模擬 Java 版的重連與主題重整邏輯
		_tick := 0
		for {
			if _ms.MQTTClient.IsConnected() {
				break
			}
			time.Sleep(1 * time.Second)
			_tick++
			if _tick >= 5 {
				_tick = 0
				// 此處依賴底層自動重連，或手動觸發連線
			}
		}

		if _ms.MQTTClient.IsConnected() {
			_ms.impl.OnMQTTConnected()

			// 訂閱必要主題
			_ms.MQTTClient.Subscribe(_ms.MQTT_Default_Topic, 0)
			_ms.MQTTClient.Subscribe(_ms.MQTT_AsyncTask_Topic, 0)

			if _ms.MQTT_Default_Topic != _topic {
				_ms.MQTTClient.Subscribe(_topic, 0)
			}

			_ms.MQTT_Topic = _topic
			Tools.Log.Print(Tools.LL_Debug, "MQTT current Topic : "+_ms.MQTT_Topic)
		}

		Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("MQTT connection status : %v", _ms.MQTTClient.IsConnected()))
	}()
}

//-------------------------------------------------------------------------------------
// 系統命令處理 (MQTT)
//-------------------------------------------------------------------------------------

// OnDefaultMQTTMessage 處理系統預設命令
func (_ms *MarsService) OnDefaultMQTTMessage(_topic, _payload string) {
	_msgObj := MarsJSON.NewJSONObject(_payload)
	// 取得 values 陣列中的第一個指令
	_values := _msgObj.OptJSONArray("values")
	if _values != nil && _values.Length() > 0 {
		_cmdObj := _values.OptJSONObject(0)
		_cmd := _cmdObj.OptString("cmd", "")

		switch _cmd {
		case "reboot":
			_ms.RestartService()
		case "shutdown":
			_ms.ShutdownService()
		case "reset_properties":
			_ms.ModifyProperties(_cmdObj.OptString("properties", ""))
		}
	}
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) ModifyProperties(_payload string) {
	if _payload != "" {
		// 儲存新的設定檔
		os.WriteFile(_ms.PropertyFileName, []byte(_payload), 0644)
		Tools.Log.Print(Tools.LL_Info, "Properties updated, restarting...")
		_ms.RestartService()
	}
}

//-------------------------------------------------------------------------------------
// HTTP 服務管理
//-------------------------------------------------------------------------------------

// ResetWebService 初始化或更新 Web 服務設定
func (_ms *MarsService) ResetWebService() {
	_method := _ms.Property.OptString("web_thread_method", "fixed")
	_coreSize := _ms.Property.OptInt("web_core_count", 0)
	_maxSize := _ms.Property.OptInt("web_thread_count", 500)
	_timeout := _ms.Property.OptInt("web_thread_timeout", 500)

	if _ms.HttpService != nil {
		// 如果服務已存在，更新其 Executor 設定
		_ms.HttpService.InitExecutor(true, _method, _coreSize, _maxSize, _timeout)

	} else {
		// 建立新的 HttpService 實例
		_ms.HttpService = HttpService.Create(
			_ms.defaultHttpPort,
			_ms.defaultHttpsPort,
			_ms.ssl_Key_File,
			_ms.ssl_Key_Password,
		)
		// 初始化執行緒池邏輯 (在 Go 中主要為設定參數)
		_ms.HttpService.InitExecutor(false, _method, _coreSize, _maxSize, _timeout)
	}

	Tools.Log.Print(Tools.LL_Debug, fmt.Sprintf("Reset Web Service : %d/%d", _ms.defaultHttpPort, _ms.defaultHttpsPort))
	//Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("Set Web Thread : %s -> %d/%d", _method, _coreSize, _maxSize))
}

// -------------------------------------------------------------------------------------
// AddRestfulAPI 動態增加 API 路由
func (_ms *MarsService) AddRestfulAPI(_uri string, _callback HttpService.HttpAPI_Callback) {

	if _uri != "" && _callback != nil {
		if _ms.HttpService != nil {
			_ms.HttpService.AddRestfulAPI(_uri, _callback)
		}
	}
}

// -------------------------------------------------------------------------------------
// RemoveRestfulAPI 移除指定的 API 路由
func (_ms *MarsService) RemoveRestfulAPI(_uri string) {
	if _uri != "" {
		if _ms.HttpService != nil {
			_ms.HttpService.RemoveRestfulAPI(_uri)
		}
	}
}

//-------------------------------------------------------------------------------------
// Web 輔助工具
//-------------------------------------------------------------------------------------

// GetWebHook 獲取當前 Web 服務的 Hook 地址
func (_ms *MarsService) GetWebHook() string {

	_ip := _ms.ServiceInfo.OptString("ip", "")

	if _ms.webHook != "" {
		return strings.Replace(strings.Replace(_ms.webHook, "127.0.0.0", _ip, 1), "localhost", _ip, 1)
	}

	return fmt.Sprintf("http://%s:%d", _ip, _ms.defaultHttpPort)
}

// ------------------------------------------------------------------------------------
func (_ms *MarsService) GetProperty() *MarsJSON.JSONObject {
	if _ms.Property != nil {
		return _ms.Property
	}

	return MarsJSON.NewJSONObject("{}")
}

// ------------------------------------------------------------------------------------
func (_ms *MarsService) MergePropertyFrom(_ext *MarsJSON.JSONObject) {
	if _ms.Property != nil {
		_ms.Property.MergeFrom(_ext)
	}
}

// ------------------------------------------------------------------------------------
func (_ms *MarsService) OnUpdateProperty() {

	if _ms.impl != nil {

		_ms.impl.OnPropertyChange(_ms.Property)
	}
}

// ------------------------------------------------------------------------------------
// SendRespone 靜態工具的服務層包裝
func (_ms *MarsService) SendRespone(_w http.ResponseWriter, _no int, _contentType string, _content []byte) {
	HttpService.SendRespone(_w, _no, _contentType, _content)
}

//-------------------------------------------------------------------------------------
// 資源管理 (GC / WebService)
//-------------------------------------------------------------------------------------

// ResetAutoGC 設定自動記憶體回收
func (_ms *MarsService) ResetAutoGC() {
	_gcInterval := _ms.Property.OptInt("auto_gc", 1800) // 預設 30 分鐘
	if _gcInterval > 0 {
		_ms.DefaultTimer = time.NewTicker(time.Duration(_gcInterval) * time.Second)
		go func() {
			for range _ms.DefaultTimer.C {
				runtime.GC()
				Tools.Log.Print(Tools.LL_Debug, "System GC executed")
			}
		}()
	}
}

// -------------------------------------------------------------------------------------
// 重啟管理
// -------------------------------------------------------------------------------------

func (_ms *MarsService) RestartService() {
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) ResetAutoRestart() {
}

//-------------------------------------------------------------------------------------
// 功能包裝 (Wrappers)
//-------------------------------------------------------------------------------------

func (_ms *MarsService) PutData(_uuid, _suid string, _data *MarsJSON.JSONObject) bool {
	if _ms.MarsClient != nil {
		return _ms.MarsClient.PutData(_uuid, _suid, _data)
	}
	return false
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) PushLineMessage(_target, _msg string) bool {
	if _ms.MarsClient != nil {
		//return _ms.MarsClient.PushLineMessage(_target, _msg)
	}
	return false
}

//-------------------------------------------------------------------------------------
// 關閉邏輯
//-------------------------------------------------------------------------------------

func (_ms *MarsService) StopService() bool {
	_ms.CloseNetService()
	Tools.Log.Print(Tools.LL_Info, "Service Stoped")
	return true
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) ShutdownService() {
	_ms.StopService()
	os.Exit(0)
}

// -------------------------------------------------------------------------------------
func (_ms *MarsService) CloseNetService() {
	if _ms.MQTTClient != nil {
		_ms.MQTTClient.Disconnect(250)
	}
	Tools.Log.Print(Tools.LL_Info, "Network services closed")
}

//-------------------------------------------------------------------------------------
