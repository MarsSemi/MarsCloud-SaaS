
## 說明

本目錄提供之 JavaScript，主要是應用於網頁端使用，
無法確認或保證可以正常運行在 node.js 底下。

## 目錄

- **第一章 : 登入**
- **第二章 : 資料存取**
- **第三章 : MQTT**
- **第四章 : 微服務功能呼叫**
- **第五章 : 進階使用**

## 使用內容

### 第一章 登入系統  

JS 的 SDK 已定義了一個全域變數 _User, 來當成使用的實例，  
無須另行宣告。其中登入系統，主要是呼叫 DoLogin 這個函式。這個  
函式包含了三個參數：帳號(user)、密碼(password)、專案(project)  

```
_User.DoLogin("test", "test", "justtest");
```

其中的專案代碼，需由系統架設人員提供，才能正確  
登入。登入之後，該登入的有效時限為 24小時，超過  
之後，需要重新進行登入，才能持續運作。
  
### 第二章 如何存取資料  

資料的取得，主要分成三個部分。

#### 2.1 資料列表

```
_User.GetUserDataSrcList();
```

呼叫本函式後，會取得一系列的列表。有列出來的部分，就代表該帳號  
可以存取的資料 Table。資料 Table，通常由 UUID、SUID 兩部分  
組成，根據各專案需求自行命名。需注意的是，字母的大小寫不同，會  
被視為不同的名稱，請特別留意。以習慣來說，UUID 通常會是 資料    
的類別、IoT 設備的編號、或是第三來源端的識別碼。如以下範例：  

```
lorabox.tempmetter
```
其中 lorabox 為 UUID 代表就的是類型，tempmeter 為 SUID  
表示的就是溫度量測計。這是一種命名方式，另外也可以變成這樣：  

```
E0ABDDEC9C.info
E0ABDDEC9C.data
E0ABDDEC9C.report
```
上述例子，則以 IoT 設備的硬體編號為 UUID，SUID 為描述，  
來設計資料分割的結構。當然反過來也是可以的：  

```
info.E0ABDDEC9C
data.E0ABDDEC9C
report.E0ABDDEC9C
```
這種設計思維，UUID 則是以資料類型來分隔，然後再用設備碼  
當 SUID 來做管理。  

不同的設計有不同的優缺點，要視該專案適合的類型，再來決定  
資料切割的方式。沒有一定的答案，只有適不適合的問題要考量。  
如果不是很清楚該怎麼設計，歡迎來信或來電詢問本公司，會有  
專門的顧問，來回答、分析你的需求。

#### 2.2 資料寫入與取得

寫入資料時，需指定資料要寫入哪個 Table，也就是要指定 UUID  
與 SUID。並盡量以 JSON 的格式傳入，以利後續修改資料內容：  

```
let _data =
{
  temp: 23.5,               
  humi: 70.2,
  temp_unit: "°C", 
  humi_unit: "%"
};

_User.updateDataByKey(null, "UUID.SUID", [data key, can be null], _data);

```
上述範例，其中第三個參數，為該筆資料的 UKey，可以是 null。  
當 UKey 設定為 null 時，系統會自動填上一個流水編號，若要特  
別指定則自行填入。相同 UKey 的資料會被覆蓋，需要特別注意這點。  

資料的取得部分，大致上有三種方式：  

- **getData      : 使用時間區間取得資料**
  
```
getData 的參數意義如下：
  
_User.getData("UUID.SUID", [start_time], [end_time], [callback, can be null], [user define item, can be null]);  

其中的時間區間，為系統時間，以 million second 為單位。
CallBack Function 有設定時，則會以 non-block 方式
運作，大量資料存取時，可以讓網頁不卡死。

function MyCallBack(_data, _myitem){ ... }  
...  
...  
  
_User.getData("UUID.SUID", 195464870000, 19546540000, MyCallBack, _myitem);  

```
- **getDataByKey : 以指定UKey的方式取得資料**

```
getDataByKey 的參數意義如下：
  
_User.getDataByKey("UUID.SUID", "data_ukey", [callback, can be null], [user define item, can be null]);  

除了改成使用 Ukey 外，其餘參數與 getData 相同。
```
  
- **getLastData  : 以時間倒序的方式取得資料**
  
```
getLastData 的參數意義如下：
  
_User.getLastData("UUID.SUID", "data_count", [callback, can be null], [user define item, can be null]);  

除了指定資料 count 外，其餘參數與 getData 相同。
當 count 設為 0 時，代表要取出該 Table 所有的資料。
如果資料量很大，會造成網頁速度變慢或佔用記憶體，請謹慎
使用。
```
  
### 第三章 MQTT 與訊息收發  

JS 的 SDK，引用了 paho-mqtt 的 MQTT Client。本原生雲  
系統，目前支援 3.X 的 MQTT 協議。系統中有三種類的 MQTT  
訊息，分別是 data、event、自訂。Data 為任何資料更新時，
都會收到的訊息，使用方法如下：  

```
function OnDataChange(_msg)
{
}

範例一 ： _User.SubscribeData('*', OnDataChange); //subscibe all
範例二 ： _User.SubscribeData('dev.E2F0A9C3B5', OnDataChange); //subscibe specify data
```

以上述範例來說，當有台裝置 dev.AABBCCDD 更新或新增資料時，
若是全部訂閱的狀況，會收到該筆資料更新。而訂閱 dev.E2F0A9C3B5  
的時候，則不會收到該筆資料更新。

而 event 訂閱，則是會收到呼叫系統 Put Event 功能所拋出的  
事件，如下：

```
function OnPushEvent(_msg)
{
}

範例一 ： _User.SubscribeEvent('*', OnPushEvent); //subscibe all
範例二 ： _User.SubscribeEvent('dev.E2F0A9C3B5', OnPushEvent); //subscibe specify data
```
自定義的方式，則是透過呼叫 SubscribeByMQTT_Adv 函式來完成，  
其中的 Topic 就是對應 MQTT 中的 Topic, 使用起來也完全相同。  
而要推送自定義的訊息，則是須透過呼叫系統的 Put Message 功能  
來完成。
  
### 第四章 微服務功能呼叫  
  
### 第五章 進階使用  
  
