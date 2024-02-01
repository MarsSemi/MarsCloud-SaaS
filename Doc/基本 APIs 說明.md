 
## 快速 Restful APIs 使用

本章節提供開發人員快速銜接系統的相關 API，讓整合的時間縮  
短。如需進一步的詳細資訊，請移至本文的最後，並點開相關鏈結，  
來獲得更詳細的使用說明。
  
### 01. 快速登入
  
<table>
  <tr>
    <td>項目</td>
    <td>基本登入</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/login</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Post</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{"usr": "test", "pwd": "justtest", "proj": "justtest" }</td>
  </tr>
</table>
  
使用 POST 功能，並傳送正確的內容。呼叫成功後，會返回一 token 字串如下，  
  
```
eyJjb20iOiJtYXJzLXNlbWkuY29tIiwiYWxnIjoiZGlyIiwiZW5jIjoiQTEyOEdDTSJ9.XX ...
```
  
用於後續的資料、功能的存取。該 token 字串，會在24小時之後，失去使用  
授權。此時必須重新進行登入，獲取新的 token 來保持系統正常的運作。  
  
其中，帳號與密碼為基本需求資料。而 proj 參數，則是必須要指定的專案名稱。  
本雲端系統，允許客戶在同一雲端服務下，建置不同的專案運作。而不同的專案，  
其資料、設置、微服務等，是獨立運作，彼此互不干擾的。因此可實現，單一平台，  
服務多個專案的功能，讓資源得到極大化的使用。

  
### 02. 資料存取

#### 寫入一筆資料
  
本雲端系統的資料儲存，是以 UUID 與 SUID 兩個參數來組成該資料的 Table，  
通常 UUID 會使用資料的類別來命名，而 SUID 則是該類別底下的子項目。  
以下是存入一筆資料的基礎範例：

<table>
  <tr>
    <td>項目</td>
    <td>資料寫入</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/put?data</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Post</td>
  </tr>
  <tr>
    <td>Headers</td>
    <td>Authentication : Bearer [login token]</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{"uuid": "employee", "suid": "member", "values": [{"ukey": "unique_id", "key1": "value1", "key2": 001, "key3": true}] }</td>
  </tr>
</table>

在上述的例子中，我們存入了一筆 ukey 值為 unique_id 的資料，  
其內容為 values 裡的一項 JSONObject 格式。而 ukey 就是該  
筆資料的唯一識別碼，用來存取這筆資料。若是存入資料時，沒有指定  
ukey，則系統會依據時間自行給個流水號碼作為該筆資料的識別碼。  

如果要上傳多筆資料，請不要重複呼叫這個 API 來上傳，照樣會導致  
效率低落。而是需要在 values 這個 JSONArray 中，放入多筆資料。  
但也不要上傳無限制的數量，一來說呼叫一次 API，上傳資料筆數建議  
是 3000~8000 筆內比較合適，依主系統的運算能力來決定。


#### 讀出一筆資料

<table>
  <tr>
    <td>項目</td>
    <td>指定資料讀出</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/get?data</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Post</td>
  </tr>
  <tr>
    <td>Headers</td>
    <td>Authentication : Bearer [login token]</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{ "uuid": "employee", "suid": "member", "ukey": "unique_id" }</td>
  </tr>
</table>
  
指定 UUID、SUID 與 Ukey 來取得單筆的指定資料。  

#### 讀出最後的數筆資料

<table>
  <tr>
    <td>項目</td>
    <td>最後資料讀出</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/lastdata?method=read</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Post</td>
  </tr>
  <tr>
    <td>Headers</td>
    <td>Authentication : Bearer [login token]</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{ "uuid": "employee", "suid": "member", "count": 5 }</td>
  </tr>
</table>

上述的例子，指定了 UUID、SUID 與 Count，來取得該資料叢集的最後5筆資料。  

#### 讀出指定日期的資料

<table>
  <tr>
    <td>項目</td>
    <td>指定日期資料讀出</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/getbyday</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Post</td>
  </tr>
  <tr>
    <td>Headers</td>
    <td>Authentication : Bearer [login token]</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{ "uuid": "employee", "suid": "member", "utc_time": 1896... }</td>
  </tr>
</table>

上述的例子，指定了 UUID、SUID 與 日期，來取得該資料叢集指定日期的資料。  
這邊需要特別注意的是，utc_time 這個參數，使用的是 UTC 時間，記得加上  
所在地域的 TimeZone Offset, 輸入的單位為秒。  
  
#### 移除指定資料

<table>
  <tr>
    <td>項目</td>
    <td>指定資料刪除</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/del?data</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Post</td>
  </tr>
  <tr>
    <td>Headers</td>
    <td>Authentication : Bearer [login token]</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{ "uuid": "employee", "suid": "member", "ukey": "unique_id" }</td>
  </tr>
</table>

上述的例子，指定 UUID、SUID 與 Ukey 來刪除單筆的指定資料。  
    
### 03. 可用資料列表


<table>
  <tr>
    <td>項目</td>
    <td>取得資料叢集列表</td>
  </tr>
  <tr>
    <td>API</td>
    <td>https://test.mars-cloud.com/api/usrinfo?method=datasrcinfo</td>
  </tr>
  <tr>
    <td>Method</td>
    <td>HTTP Pㄔㄟ</td>
  </tr>
  <tr>
    <td>Headers</td>
    <td>Authentication : Bearer [login token]</td>
  </tr>
  <tr>
    <td>Content</td>
    <td>{ "uuid": "employee", "suid": "member" }</td>
  </tr>
</table>

在第二章節，我們敘述了資料存取的方式，並指定了 UUID 與 SUID 來  
獲取指定叢集的內容。但在某些情況下，我們必須判斷資料叢集是否存在。  
因此使用上述的指令，就可以獲得如下範例的列表。

```
{
        "uuid": "Sensor",
        "suid": "B827EBD211B9",
        "name": "Sensor",
        "ttl": 5184000,
        "data_profile": "both.Sensor",
        "vender_id": "virtual.com",
        "owner": "",
        "record_count": 3182,
        "share_key": "",
        "record_size": 428615,
        "share_loca": "",
        "type": "",
        "product_id": "",
        "forceAlarm": "0",
        "ukey": "Sensor_B827EBD211B9",
        "ext3": "",
        "ext2": "",
        "ext1": "",
        "desc": ""
}
```
    
### 04. 呼叫微服務 API


  
## 全部 Restful APIs 相關資料
  
詳細內容 : https://www.mars-cloud.com/portal/api/api_document.html    
