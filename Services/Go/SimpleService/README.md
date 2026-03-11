# Simple Service

一個基於 [MarsCloud SDK](https://github.com/MarsSemi/MarsCloud-SaaS/SDK/Go) 開發的微服務範例專案。本專案展示了如何整合 MQTT 通訊、RESTful API 服務以及定時任務。

## 🚀 功能特點

- **微服務框架整合**：使用 `MarsService` 進行快速開發與 MarsCloud 註冊。
- **定時任務**：在 `MyCloudService.Process` 中實作背景計數器，每 30 秒記錄一次系統狀態。
- **MQTT 通訊**：自動處理 MQTT 連線、斷連回呼，並提供訊息解析範例（`OnMQTTMessage`）。
- **RESTful API**：內建 API 處理機制，範例提供 `/api/hello` 端點，支援 JWT 驗證。
- **異常處理**：整合 `Tools` 模組的 UncaughtExceptionHandler 與日誌系統。

## 📂 目錄結構

- `main.go`: 程式進入點，負責初始化日誌與服務生命週期。
- `MyService.go`: 定義 `MyCloudService` 結構，包含主要的背景邏輯與 MQTT 事件處理。
- `HttpAPI_Api.go`: 實作 RESTful API 的業務邏輯（`/api` 路徑）。
- `agent.properties`: 服務配置檔，定義雲端連線資訊、通訊埠與重啟策略。
- `go.mod`: 專案依賴管理文件。

## 🛠️ 本地開發

### 環境需求
- Go 版本：1.25.6 或更高。

### 安裝步驟
1. 複製專案。
2. 下載依賴封包：
   ```bash
   go mod download
   ```

### 執行服務
你可以直接透過編譯執行：
```bash
go run .
```

或是透過腳本執行：
```bash
# 基本啟動
./run.sh

# 完整編譯與啟動
./runFull.sh
```

---
*詳細部署細節請參考 [DEPLOY.md](DEPLOY.md)*
