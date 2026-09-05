# MarsService

`MarsService` 是整個 SDK 的 service 框架核心，整合 HTTP/HTTPS、MarsCloud 登入、MQTT client、本地 MQTT broker、設定檔管理與 service lifecycle。

## 主要能力

- 根據 `agent.properties` 建立 service
- 啟動 HTTP / HTTPS server
- 連線 MarsCloud 並建立 MQTT client
- 註冊 service 與 properties
- 管理自動重啟、GC、關機清理
- 在需要時啟動本地 MQTT broker

## 主要型別

- `IMarsService`
- `MarsService`

## `IMarsService` 介面

```go
type IMarsService interface {
    OnMQTTConnected()
    OnMQTTMessage(_topic string, _payload string)
    OnMQTTConnectionLost(_err error)
    OnPropertyChange(*MarsJSON.JSONObject)
    BeforeServiceStop()
    Process()
}
```

## 常用函式

- `Create(_propertyFileName string, _impl IMarsService) *MarsService`
- `(_this *MarsService) Start()`
- `(_this *MarsService) StopService() bool`
- `(_this *MarsService) RestartService()`
- `(_this *MarsService) ShutdownService()`
- `(_this *MarsService) AddRestfulAPI(_uri string, _callback HttpService.HttpAPI_Callback)`
- `(_this *MarsService) RegistryServerInfo(_version string, _type string, _isOnline bool)`
- `(_this *MarsService) SetLocalMQTTMessageCallback(_callback MQTTServer.MessageCallback)`
- `(_this *MarsService) SendResponse(...)`

## 啟動流程摘要

1. 讀取 `agent.properties`
2. 初始化 `HttpService`，完成 HTTP/HTTPS listener 綁定與 TLS 憑證載入
3. 視設定啟動本地 MQTT broker
4. 監聽埠、TLS 或 broker 初始化失敗時記錄錯誤並以狀態碼 `1` 退出，讓外部管理器依策略重試
5. 若 `mars_cloud_url/account/password` 完整，在背景連線 MarsCloud；連線失敗時持續重試，不阻塞 HTTP/HTTPS 啟動
6. 啟動重啟排程與 GC；MarsCloud 連線成功後建立 MQTT client、AsyncTaskProcessor 並執行 service registry

## 一般範例

```go
type MyService struct {
    *MarsService.MarsService
}

func (s *MyService) OnMQTTConnected() {}
func (s *MyService) OnMQTTMessage(topic, payload string) {}
func (s *MyService) OnMQTTConnectionLost(err error) {}
func (s *MyService) OnPropertyChange(prop *MarsJSON.JSONObject) {}
func (s *MyService) BeforeServiceStop() {}
func (s *MyService) Process() {}

func main() {
    svc := &MyService{}
    ms := MarsService.Create("agent.properties", svc)
    ms.SetLocalMQTTMessageCallback(func(topic, payload string) {})
    ms.RegistryServerInfo("1.0.0", "pack", true)
    ms.Start()
    select {}
}
```

## 關鍵設定

- `mars_cloud_url`
- `mars_cloud_account`
- `mars_cloud_password`
- `mars_cloud_proj`
- `mqtt_server_enable`
- `mqtt_allow_anonymous`
- `mqtt_username`
- `mqtt_password`
- `http_port`
- `https_port`
- `ssl_key`
- `ssl_key_file`
- `ssl_key_password`
- `restart_time`：定時自動重啟時間清單，例如 `["06:00", "14:30"]`
- `restart_timezone`：`restart_time` 比對所用時區（IANA 名稱，如 `Asia/Taipei`、`UTC`）；未設定或解析失敗時退回 `time.Local`

## 注意事項

- `mqtt_server_enable` 預設為 `false`
- 本地 MQTT broker 預設拒絕匿名連線；啟用 broker 時須設定 `mqtt_username` 與 `mqtt_password`，只有明確設定 `mqtt_allow_anonymous=true` 才允許匿名連線
- HTTP/HTTPS Server 不需等待 MarsCloud 連線成功；MarsCloud 無法連線時，仍可提供不依賴雲端資源的 HTTP API 與靜態檔案
- 內建 `/system` 管理 API 只有在 SDK token 驗證成功且 claims 非空時才會執行；未驗證請求回傳 `401 Unauthorized`
- 缺少任一 MarsCloud 登入欄位時，只會當一般 server 啟動，不會建立雲端 MQTT client、AsyncTaskProcessor 或執行 service registry
- `Start()` 會以 goroutine 非同步執行啟動流程；呼叫返回不代表 HTTP/HTTPS 已開始監聽，主程式需保持運行
- 依賴 `MarsClient`、雲端 `MQTTClient` 或 `AsyncTaskProcessor` 的 API，應在使用前確認元件已完成初始化或雲端已連線
- `OnMQTTMessage` 是雲端 MQTT client 的回調，本地 broker 則用 `SetLocalMQTTMessageCallback`
- 啟動時不再關閉同名或佔用連接埠的程序；連接埠衝突會回報啟動失敗。舊設定 `conflict_restart` 不再用來清除程序或反覆重啟
- 啟動時會 log 出當前 `Restart Timezone`，遠端容器若 `/etc/localtime` 缺失而 fallback `UTC` 可立即看出
- 同一實例只啟動一次；重啟要求使用互斥控制，已有啟停操作時略過重複要求

## 部署注意：與 systemd 的相容性

Unix（含 Linux、macOS）以 `exec` 直接替換目前程序，保留 PID、工作目錄、參數與環境。`exec` 回傳錯誤時，原服務的 listener 與背景工作繼續運作。systemd 的 `Type=simple` 能繼續追蹤相同 PID，內建重啟不需要改成 `Type=forking`。

Unix 內建重啟不呼叫 `BeforeServiceStop()` 或雲端離線通知，以免在 `exec` 失敗時無法恢復已清理的業務資源。作業系統會在程序替換時關閉一般 Go 網路連線；正在處理的請求會中斷。需要先完成業務清理時，應使用停止訊號並由外部管理器重新啟動。

Windows 直接建立新程序，不經 `cmd start`。新程序透過繼承的父程序 handle 等待舊程序退出，再繼續初始化；建立失敗時原服務繼續運作。成功建立後，舊程序執行停止清理並退出，清理最多等待 30 秒。此方式適用於一般程序啟動，不代表 Windows 服務管理器會自動接管新 PID。

成功替換或建立程序只代表作業系統接受啟動，不代表後續業務初始化完成。HTTP/HTTPS、TLS 或 broker 初始化失敗時會以狀態碼 `1` 退出；正式部署應搭配 systemd 的 `Restart=on-failure`、適當的 `RestartSec` 或其他外部管理器，處理啟動後失敗。此機制不提供零中斷或啟動後自動回復舊版本的保證。
