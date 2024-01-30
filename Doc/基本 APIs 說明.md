 
## 01. Restful APIs 使用 : 快速登入
  
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
  
使用 POST 功能，並傳送下列內容。呼叫成功後，會返回一串 token 字串，  
用於後續的資料、功能的存取。該 token 字串，會在24小時之後，失去使用  
授權。此時必須重新進行登入，獲取新的 token 來保持系統正常的運作。  
  
其中，帳號與密碼為基本需求資料。而 proj 參數，則是必須要指定的專案名稱。  
本雲端系統，允許客戶在同一雲端服務下，建置不同的專案運作。而不同的專案，  
其資料、設置、微服務等，是獨立運作，彼此互不干擾的。因此可實現，單一平台，
服務多個專案的功能，讓資源得到極大化的使用。


## 全部 Restful APIs 相關資料
  
詳細內容 : https://www.mars-cloud.com/portal/api/api_document.html    
