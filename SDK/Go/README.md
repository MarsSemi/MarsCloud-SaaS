# MarsCloud-SaaS Go SDK

MarsCloud-SaaS SDK for Go，提供 MarsCloud 雲端服務的開發套件。

## 版本

- **SDK 版本**: v0.1.12
- **Go 版本**: 1.18+

## 安裝

```bash
go get github.com/MarsSemi/MarsCloud-SaaS/SDK/Go
```

## 模組總覽

| 模組 | 說明 |
|------|------|
| MarsService | 主服務框架，整合 MQTT、HTTP、WebSocket |
| MarsClient | MarsCloud 客戶端，處理認證、訊息收發 |
| MarsJSON | JSON 處理工具，支援 JSONObject 與 JSONArray |
| MQTTClient | MQTT 客戶端，訊息訂閱與發佈 |
| HttpService | HTTP 服務框架，RESTful API 處理 |
| Security | 安全模組，包含 JWT、AES、RSA 加密 |
| Tools | 工具集合，日誌、網路、檔案操作 |
| DataCompress | 資料壓縮，ZIP 加密/解密 |
| AsyncTaskProcessor | 非同步任務處理器 |
| ScriptExecutor | JavaScript 腳本引擎 |

---

## 模組詳解

### MarsService

主服務框架，提供 MarsCloud 服務的完整解決方案。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsService"
```

**主要功能**：
- MQTT 連線管理
- HTTP/HTTPS 伺服器
- WebSocket 支援
- 屬性配置管理
- 訊息處理

**基本使用**：
```go
type MyService struct {
    *MarsService.MarsService
    Counter int
}

func (s *MyService) OnMQTTConnected() {
    Tools.Log.Print(Tools.LL_Info, "MQTT connected")
}

func (s *MyService) OnMQTTMessage(topic, payload string) {
    Tools.Log.Print(Tools.LL_Info, "Received: %s", payload)
}

func main() {
    service := &MyService{Counter: 0}
    ms := MarsService.Create("agent.properties", service)
    ms.Start()
}
```

---

### MarsClient

MarsCloud 客戶端，負責與 MarsCloud 伺服器通訊。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsClient"
```

**主要功能**：
- 連線管理
- 訊息發佈/訂閱
- 檔案傳輸（Base64 編碼）
- 屬性同步

---

### MarsJSON

JSON 處理工具，類似 JavaScript 的物件操作風格。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
```

**主要功能**：
- JSONObject 鍵值操作
- JSONArray 陣列操作
- 自動類型轉換

**基本使用**：
```go
// 解析 JSON
obj := MarsJSON.NewJSONObject(`{"name": "test", "value": 123}`)

// 獲取值
name := obj.OptString("name", "default")
value := obj.OptInt("value", 0)

// 設置值
obj.Set("newKey", "newValue")

// 轉為字串
jsonStr := obj.ToString()
```

---

### MQTTClient

MQTT 訊息客戶端，基於 Eclipse Paho MQTT 庫。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MQTTClient"
```

**主要功能**：
- QoS 0/1/2 支援
- 訊息訂閱與發佈
- 保留訊息
- 連線狀態回調

---

### HttpService

HTTP 服務框架，簡化 RESTful API 開發。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/HttpService"
```

**主要功能**：
- 路由處理
- 靜態檔案服務
- WebSocket 升級
- CORS 支援

---

### Security

安全相關功能，包含加密與認證。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Security"
```

**主要功能**：
- **JWT**: JSON Web Token 產生與驗證
- **AES**: AES 加密/解密
- **RSA**: RSA 加密/解密
- **UserVerify**: 用戶認證與權限管理

---

### Tools

工具集合，包含多種常用功能。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
```

**主要功能**：
- **Log**: 日誌系統，支援多級別輸出
- **HTTP**: HTTP 客戶端工具
- **Net**: 網路工具
- **File**: 檔案操作
- **Encode**: Base64 編碼
- **Mail**: SMTP 郵件發送
- **Time**: 時間處理

**日誌級別**：
```go
Tools.Log.SetDisplayLevel(Tools.LL_Debug)   // 調試
Tools.Log.SetDisplayLevel(Tools.LL_Normal)  // 一般
Tools.Log.SetDisplayLevel(Tools.LL_Info)    // 資訊
Tools.Log.SetDisplayLevel(Tools.LL_Warning) // 警告
Tools.Log.SetDisplayLevel(Tools.LL_Error)   // 錯誤
```

---

### DataCompress

資料壓縮工具。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/DataCompress"
```

**主要功能**：
- ZIP 檔案解壓
- 加密 ZIP 支援

Processor

非同步---

### AsyncTask任務處理器。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/AsyncTaskProcessor"
```

**主要功能**：
- 非同步任務排程
- Webhook 回調

---

### ScriptExecutor

JavaScript 腳本引擎，基於 goja。

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/ScriptExecutor"
```

**主要功能**：
- 執行 JavaScript 腳本
- 腳本熱重載
- 預編譯腳本

---

## 範例程式

### 基本伺服器

```go
package main

import (
    "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsService"
    "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/MarsJSON"
    "github.com/MarsSemi/MarsCloud-SaaS/SDK/Go/Tools"
)

type MyService struct {
    *MarsService.MarsService
}

func (s *MyService) OnMQTTConnected() {
    Tools.Log.Print(Tools.LL_Info, "Connected to MarsCloud")
}

func (s *MyService) OnMQTTMessage(topic, payload string) {
    Tools.Log.Print(Tools.LL_Info, "Message: %s -> %s", topic, payload)
}

func (s *MyService) OnPropertyChange(prop *MarsJSON.JSONObject) {
    Tools.Log.Print(Tools.LL_Info, "Property changed")
}

func main() {
    service := &MyService{}
    ms := MarsService.Create("agent.properties", service)
    ms.Start()
}
```

---

## 依賴套件

SDK 依賴以下第三方套件：

- `github.com/eclipse/paho.mqtt.golang` - MQTT 客戶端
- `github.com/gorilla/websocket` - WebSocket
- `github.com/dop251/goja` - JavaScript 引擎
- `github.com/lestrrat-go/jwx` - JWT/JWE/JWS
- `github.com/alexmullins/zip` - 加密 ZIP

完整依賴請參考 [go.mod](https://github.com/MarsSemi/MarsCloud-SaaS/blob/main/SDK/Go/go.mod)。

---

## 授權

本 SDK 遵循 MarsCloud 服務條款。

---

## 相關連結

- [MarsCloud 官網](https://mars-cloud.com)
- [GitHub 倉庫](https://github.com/MarsSemi/MarsCloud-SaaS)
