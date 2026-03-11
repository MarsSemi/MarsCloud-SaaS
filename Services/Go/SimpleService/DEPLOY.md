# 部署手冊 (DEPLOY.md)

本文件說明如何將 `Simple Service` 部署至生產環境，並對 `agent.properties` 配置進行詳細說明。

## 📋 配置說明 (agent.properties)

服務啟動時會讀取專案目錄下的 `agent.properties` JSON 檔案：

| 參數名 | 說明 | 範例 |
| :--- | :--- | :--- |
| `service_name` | 顯示於雲端管理介面的服務名稱 | `"Simple Service"` |
| `mars_cloud_url` | MarsCloud 伺服器之連線地址 | `"https://test.mars-cloud.com"` |
| `mars_cloud_account` | 雲端帳號 | `"test"` |
| `mars_cloud_password` | 雲端密碼 | `"test"` |
| `mars_cloud_proj` | 所屬專案名稱 | `"justtest"` |
| `http_port` | HTTP 服務通訊埠 | `80` |
| `https_port` | HTTPS 服務通訊埠 | `443` |
| `restart_time` | 服務自動定時啟動的時間點 (24小時制) | `["06:00:00"]` |

## 🚀 部署步驟

### 1. 編譯二進位檔案
在 Linux 或 macOS 環境下，使用以下指令進行編譯：
```bash
go build -o SimpleService .
```

### 2. 環境配置
確保伺服器上具備以下檔案：
- `SimpleService` (執行檔)
- `agent.properties` (配置檔)

### 3. 啟動服務
建議透過預設腳本啟動以確保資源清理與正確載入：
```bash
chmod +x run.sh runFull.sh
./run.sh
```

## 🛠️ 維運與排錯

### 自動重啟機制
本服務整合了 MarsCloud SDK 的定時重啟功能。若在 `agent.properties` 中設定了 `restart_time`，服務將在指定時間點自動結束行程，由外部守護程序（如 Systemd 或 Supervisor）重新拉起。

### 日誌檢查
日誌預設輸出至標準輸出。若需調整日誌等級，可在 `main.go` 中修改：
```go
Tools.Log.SetDisplayLevel(Tools.LL_Info)
```

### 系統守護程序範例 (Systemd)
建立 `/etc/systemd/system/simpleservice.service`:
```ini
[Unit]
Description=Simple Service
After=network.target

[Service]
Type=simple
User=vader
WorkingDirectory=/home/vader/SimpleService
ExecStart=/home/vader/SimpleService/SimpleService
Restart=always

[Install]
WantedBy=multi-user.target
```
