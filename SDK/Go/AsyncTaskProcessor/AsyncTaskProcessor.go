package AsyncTaskProcessor

// -------------------------------------------------------------------------------------
import (
	"strings"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsClient"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
)

// -------------------------------------------------------------------------------------
//
// -------------------------------------------------------------------------------------
// AsyncTaskProcessor 處理非同步任務的處理器
type AsyncTaskProcessor struct {
	_Client            *MarsClient.MarsClient
	_MainServerType    string
	_MainServerHost    string
	_MainServerWebhook string
}

// -------------------------------------------------------------------------------------
// NewAsyncTaskProcessor 建立處理器實例
func NewAsyncTaskProcessor(_client *MarsClient.MarsClient, _type string, _webHook string) *AsyncTaskProcessor {
	_atp := &AsyncTaskProcessor{}

	if _client != nil {
		_atp._Client = _client // 注意：之前 MarsService 實作中 MarsClient 存放於 _MarsClient
		if _atp._Client != nil {
			_atp._MainServerHost = _atp._Client.GetServerURLByIndex(0)
			_atp._MainServerType = strings.ToLower(_type)
			_atp._MainServerWebhook = _webHook
		}
	}

	return _atp
}

// -------------------------------------------------------------------------------------
// OnMQTTMessage 接收並解析 MQTT 訊息
func (_atp *AsyncTaskProcessor) OnMQTTMessage(_topic string, _payload string) {
	// Go 中的 strings.Split 效果等同於 Java 的 split
	_topics := strings.Split(_topic, "/")

	// Java: if(_topics.length >= 3) switch(_topics[2])
	if len(_topics) >= 3 {
		switch _topics[2] {
		case "api":
			_atp.ProcessAPI(_payload)
		}
	}
}

// -------------------------------------------------------------------------------------
// ProcessAPI 啟動一個 Goroutine 來執行非同步任務 (對應 Java 的 new Thread().start())
func (_atp *AsyncTaskProcessor) ProcessAPI(_payload string) {
	// 在 Go 中直接使用 goroutine 處理非同步邏輯
	go func() {
		// 捕捉可能發生的 panic 以確保服務穩定
		// (這裡可以使用之前在 sysutil 實作過的 GlobalRecovery)

		var _data string = _payload

		// 1. 執行非同步任務邏輯 (對應 AsyncTask.run)
		_content := MarsJSON.NewJSONObject(_data)
		_api := _content.OptString("api", "")
		_token := _content.OptString("token", "")
		_body := _content.OptString("body", "")

		// 移除 token 不回傳
		_content.Remove("token")

		// 執行本地 HTTP POST (設定 2 小時逾時，對應 Java 2*60*60*1000)
		// 這裡呼叫之前 netutil 中支援 timeout 的方法
		_resp := Tools.HttpPost(_atp._MainServerWebhook+_api, _token, "", _body, 7200000)

		// 更新回傳內容
		_content.Put("respone", _resp)
		_data = _content.ToString()

		// 2. 回傳結果給主伺服器 (最後執行的 finally 區塊)
		if _atp._Client != nil {
			_respURL := _atp._MainServerHost + "/services/respone"
			Tools.HttpPost(_respURL, _atp._Client.AuthToken, "", _data, 15000)
		}
	}()
}

// -------------------------------------------------------------------------------------
