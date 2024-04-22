## 說明

本目錄提供 C# 的開發工具與範例，讓開發人員可以快速地開進行  
雲端系統介接的用戶端程式，進而滿足不同的用戶端需求。

## 建置環境

目前已測試過的編譯環境如下：
  
<table>
  <tr>
    <td>版本</td>
    <td>CPU</td>
    <td>OS</td>
    <td>運作結果</td>
  </tr>
  <tr>
    <td>.NET 4.X</td>
    <td>x86、x64</td>
    <td>Windows</td>
    <td>OK</td>
  </tr>
</table>
  
其中會使用 **Newtonsoft.Json** 來解析 JSON 物件，該物件  
可在以下的連結下載後使用：

[官網](https://www.newtonsoft.com/json) or [GitHub](https://www.newtonsoft.com/json)  

**Newtonsoft.Json** 下載後，記得在確認兩件事情  

- **確認對應正確的 .NET 版本**
- **確認專案有加入該 Newtonsoft.Json.dll 的參考**

接著便可正常編譯與執行
