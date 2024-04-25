## 說明

本目錄下，存放著各類微服務的範例，方便所有人快速查閱與撰寫。
  
每個微服務，皆使用 Eclipse 與 Maven 管理  
下載後直接 import 專案即可  

編譯後輸出 jar 檔案的位置為 /bin 底下  
預設的版本號以 1.YY.MMDD 的方式自動註記  
若有修改需求，可以自行至 pom.xml 裡面變更  

  
## 目錄

- **SimpleService : 最基礎的範例，可建立一個空的服務，並註冊至主系統。**

## 編譯方式

本區域的範例，皆使用 Eclipse IDE 與 MAVEN 來建置。Eclipse  
可至這裡下載：[Eclipse 官網](https://eclipseide.org/)。 
  
安裝時，請選擇 Eclipse IDE 版本即可，若有已安裝 Eclipse  
IDE for Java EE 的使用者，也可直接沿用，不需再裝回一般版。  

<img src="https://test.mars-cloud.com/images/1714029084639.jpg" width="480"></img>

安裝完後，便可開啟 Eclipse 進行 maven 專案管理的安裝。maven 專案  
管理是一個設計給 Java 專用的自動化建構工具。可以有效地建立起 Java  
檔案的管理，並有效地導入對應的依賴套件。安裝詳情可參考[這裡](https://ithelp.ithome.com.tw/articles/10303335)，本站  
不多加詳述。安裝完 maven 後，即可進行專案的導入。導入時，請在左方(預設  
位置) Projects 區塊中，按下右鍵，即可看到 imports/導入 選項：  
  
<img src="https://test.mars-cloud.com/images/1714029690471.jpg" width="480"></img>

接著透過導入，選取在本站所下載之微服務的 pom 專案檔案，就會匯入該  
微服務之專案。

<img src="https://test.mars-cloud.com/images/1714029984685.jpg" width="480"></img>  

<img src="https://test.mars-cloud.com/images/1714030139496.jpg" width="480"></img>  

匯入完成後即可進行編譯，編譯方式請不要直接執行 IDE 的 Run。請依下圖  
的方式，使用 maven install 的模式來編譯。

<img src="https://test.mars-cloud.com/images/1714028541450.jpg" width="480"></img>  

編譯完成後，就會在該目錄的 bin 檔案夾，獲得一個封裝好的 .jar 檔案。  
這個檔案便是最後需要部署在雲端伺服器的微服務執行程序。但由於本公司的  
原生雲系統，已經將混合雲的部分進行自動整合，所以也可以將此執行程序，  
運作在地端電腦中。運作方式請看參考下一章節。
  
<img src="https://test.mars-cloud.com/images/1714028621843.jpg" width="480"></img>

## 聯絡方式

- [EMail] service@mars-semi.com.tw
