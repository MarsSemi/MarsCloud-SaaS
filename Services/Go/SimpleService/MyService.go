package main

import (
	"fmt"
	"time"

	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
	"github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
)

// -------------------------------------------------------------------------------------
// MyCloudService 繼承自 MarsService
// -------------------------------------------------------------------------------------
type MyCloudService struct {
	Counter int
}

// -------------------------------------------------------------------------------------
func (_s *MyCloudService) Process() {
	Tools.Log.Print(Tools.LL_Info, "MyCloudService 主程序啟動...")
	for {
		_s.Counter++

		Tools.Log.Print(Tools.LL_Info, fmt.Sprintf("[%s] 目前計數: %d\n", time.Now().Format("15:04:05"), _s.Counter))
		time.Sleep(30 * time.Second)
	}
}

// -------------------------------------------------------------------------------------
func (_s *MyCloudService) OnMQTTConnected() {

	Tools.Log.Print(Tools.LL_Debug, "OnMQTTConnected")
}

// -------------------------------------------------------------------------------------
func (_s *MyCloudService) OnMQTTConnectionLost(_err error) {

	Tools.Log.Print(Tools.LL_Debug, "OnMQTTConnectionLost")
}

// -------------------------------------------------------------------------------------
func (_s *MyCloudService) OnMQTTMessage(_topic string, _payload string) {
	fmt.Printf("Get MQTT : %s, 內容: %s\n", _topic, _payload)

	// 範例：解析 JSON 內容並處理
	_json := MarsJSON.NewJSONObject(_payload)
	if _json.OptString("cmd", "") == "status" {
		Tools.Log.Print(Tools.LL_Info, "OnMQTTMessage")
	}
}

// -------------------------------------------------------------------------------------
func (_s *MyCloudService) OnPropertyChange(_property *MarsJSON.JSONObject) {

	Tools.Log.Print(Tools.LL_Debug, "OnPropertyChange")
}

// -------------------------------------------------------------------------------------
// BeforeServiceStop 實作關機前的清理動作
func (_s *MyCloudService) BeforeServiceStop() {
	Tools.Log.Print(Tools.LL_Debug, "My BeforeServiceStop Callback")
}

// -------------------------------------------------------------------------------------
