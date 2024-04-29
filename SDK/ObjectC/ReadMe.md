## 說明

本程式採用標準 XCcode 編譯，已測試的版本為：

<table>
  <tr>
    <td>編譯器</td>
    <td>版本</td>
    <td>CPU</td>
    <td>OS</td>
    <td>編譯結果</td>
    <td>運作結果</td>
  </tr>
  <tr>
    <td>clang</td>
    <td>15.X</td>
    <td>x64、ARM64</td>
    <td>Sonoma 14、iOS 16、iOS 17</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
  <tr>
    <td>XCode</td>
    <td>15.X</td>
    <td>x64、ARM64</td>
    <td>Sonoma 14、iOS 16、iOS 17</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
</table>

## 外部引用

本程式範例所使用的 MQTTClient 來自於 [MQTT-Client-Framework](https://github.com/novastone-media/MQTT-Client-Framework)，  
並遵循該套件的授權規則。若有進行修改與延伸使用，  
請依照該套件之規定為主。
  
## 注意事項

本範例程式 Login 後返回之 token 其時效性為  
24 Hours。需在程式中實現重新 Login 的程式，  
這個功能並不包含在本範例程式之中。
