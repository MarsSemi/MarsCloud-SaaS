# MarsCloud SaaS

MarsCloud SaaS 是 Mars Semiconductor Corp. 所提供的原生雲端與微服務整合平台。

此公開 Repository 提供產品介紹、系統架構、API 使用文件與 Go SDK，方便開發者整合 MarsCloud 服務。

## 公開內容

- `Doc/01. 原生雲設計理念.md`：原生雲設計思維與核心概念。
- `Doc/02. 原生雲設計架構.md`：微服務、REST API、MQTT 與混合雲架構。
- `Doc/03. APIs 快速使用說明.md`：MarsCloud REST API 使用方式。
- `Doc/04. 進階撰寫技巧.md`：資料識別、二進位資料與批次存取建議。
- `SDK/`：Go SDK 原始碼、Go module 與各 package 文件。

## Go SDK

Go SDK 直接位於 `SDK/`，不再使用額外的語言層級目錄。此 Repository 不處理其他程式語言的 SDK。

```bash
go get github.com/MarsSemi/MarsCloud-SaaS/SDK@v0.1.20
```

```go
import "github.com/MarsSemi/MarsCloud-SaaS/SDK/MarsService"
```

完整模組說明請參閱 [`SDK/README.md`](SDK/README.md)。SaaS 核心服務、內部工具與微服務範例仍維持於私人 Repository，不包含在本公開 Repository 中。

## 文件

- [原生雲設計理念](Doc/01.%20原生雲設計理念.md)
- [原生雲設計架構](Doc/02.%20原生雲設計架構.md)
- [APIs 快速使用說明](Doc/03.%20APIs%20快速使用說明.md)
- [進階撰寫技巧](Doc/04.%20進階撰寫技巧.md)

## 授權

本 Repository 中公開的文件與 Go SDK 依 `LICENSE` 所載條款提供。未在此 Repository 公開的 MarsCloud SaaS 核心程式、內部工具、微服務範例與相關資產不包含於本授權範圍。

如需商業授權、SDK 提前存取或技術合作，請先聯絡 Mars Semiconductor Corp.

## 聯絡方式

- Email：service@mars-semi.com.tw
- Tel：(+886) 03-5775799
