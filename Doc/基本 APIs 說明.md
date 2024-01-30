 
## 快速 Restful APIs 使用

本章節提供開發人員快速銜接系統的相關 API，讓整合的時縮短。  
如需進一步的詳細資訊，請移至本文的最後，並點開相關鏈結，  
來獲得更詳細的使用說明。
  
### 01. 快速登入
  
<table>
  <tr>
    <td>項目</td>
    <td>內容</td>
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

本雲端系統的資料儲存，是以 UUID 與 SUID 兩個參數來組成該資料的 Table，  
通常 UUID 會使用資料的類別來命名，而 SUID 則是該類別底下的子項目。  
以下是存入一筆資料的基礎範例：

<table>
  <tr>
    <td>項目</td>
    <td>內容</td>
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
    <td>Content</td>
    <td>{"uuid": "employee", "suid": "member", "values": [{"ukey": "unique_id", "key1": "value1", "key2": 001, "key3": true}] }</td>
  </tr>
</table>

在上述的例子中，我們存入了一筆 ukey 值為 unique_id 的資料，  
其內容為 values 裡的一項 JSONObject 格式。而 ukey 就是該  
筆資料的唯一識別碼，用來存取這筆資料。若是存入資料時，沒有指定  
ukey，則系統會依據時間自行給個流水號碼作為該筆資料的識別碼。  
    
### 03. 可用資料列表


    
### 04. 呼叫微服務 API


  
## 全部 Restful APIs 相關資料
  
詳細內容 : https://www.mars-cloud.com/portal/api/api_document.html    
