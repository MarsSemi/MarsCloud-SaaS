
## 說明

本目錄提供各式的開發工具或是範例，讓開發人員可以快速地開進行  
雲端系統介接的用戶端程式，進而滿足不同的用戶端需求。
  
目前列表中的開發語言，為近期內會釋出的項目。本公司的研發團隊  
會盡快地完善相關的工具與文件釋出。

## 目錄

- **Android    : 快速的 Android 開發工具**
- **Java       : 快速的應用程式開發工具**
- **Javascript : 快速的網頁端開發工具**
- **C++        : 快速的設備端開發工具**
- **C#         : 快速的應用程式開發工具**
- **ObjectC    : 快速的手機端、應用程式開發工具**
- **Swift      : 快速的手機端、應用程式開發工具**
- **Python     : 快速的應用程式開發工具**
  
## 注意事項與執行效率的提升方式

雖然範例程式已說明了使用本雲端系統的開發方式，但有些事項稍微  
注意一下，可以避免一些誤區，亦可大幅地提升整體的執行效率，以  
下就是一些常見的問題：  

#### 資料的 UKey 是什麼意思

以下是最基礎的資料上傳方式，這時資料沒有特別指定 ukey。也  
就是說，這筆資料屬於流水帳，就是一直持續會丟到資料庫裡面儲存  
的資料，常見於 IoT 的設備端 raw data。這時，系統會依據收到  
的資料時間，自動產生一個 ukey 給這筆資料，讓該資料會依照時間  
排序。

```
let _data = { temp: 23.5, humi: 70.2 };  
let _payload = { uuid: _uuid, suid: _suid, values: [_data] };

HttpPost("https://test.mars-cloud.com/api/put?data", _payload);
```

以上的這範例，最後就會產生如下的資料，存入資料庫中：

```
{ temp: 23.5, system_time: 1714704702532, humi: 70.2, ukey: 9223370322150073275_0 }
```

這時，可以看到系統會壓上一個 system time 以及一個 ukey。其  
中 system time 是系統收到該筆資料的時間，而 ukey 則為時間的  
補數（long.MAX - system_time）。一個為升冪、一個是降冪，以方  
便排序、搜尋處理。

另一種資料，則是開發者自行指定的 ukey，通常會是特定的資料內容。  
像是與帳號、設備綁定的相關資料，或是ERP、交易記錄的單號等。  

```
let _data = { ukey: _transaction_number, transaction_time: ????, transaction_result: ????,  };  
let _payload = { uuid: _uuid, suid: _suid, values: [_data] };

HttpPost("https://test.mars-cloud.com/api/put?data", _payload);
```

#### 提升儲存大量資料的效率

一般來說，儲存資料的方式大致上如上述的方式，但範例都是一筆一筆  
存入的。如果一次要存入上百筆、甚至是上萬筆資料時，如果採用單筆  
上傳，會讓效率變得極低。因此，會建議使用一次多比傳輸的方式，大  
致上的概念如下：  
  
```
let _largetData = [{ ... }, { ... }, { ... }, ..., { ... }];
let _payload = { uuid: _uuid, suid: _suid, values: _largetData };

HttpPost("https://test.mars-cloud.com/api/put?data", _payload);
```

使用以上方式，會省去多次呼叫 API 的傳輸建立時間，於是就可快速、大  
量地上傳資料。


