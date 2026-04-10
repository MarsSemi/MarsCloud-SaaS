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
2. 初始化 `HttpService`
3. 視設定決定是否啟本地 MQTT broker
4. 若 `mars_cloud_url/account/password` 完整，登入 MarsCloud
5. 建立 MQTT client、AsyncTaskProcessor、執行 service registry
6. 啟動 HTTP/HTTPS

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
- `http_port`
- `https_port`
- `ssl_key`
- `ssl_key_file`
- `ssl_key_password`

## 注意事項

- `mqtt_server_enable` 預設為 `false`
- 缺少任一 MarsCloud 登入欄位時，只會當一般 server 啟動
- `OnMQTTMessage` 是雲端 MQTT client 的回調，本地 broker 則用 `SetLocalMQTTMessageCallback`
