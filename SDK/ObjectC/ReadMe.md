## 說明

本程式採用標準 XCcode 與 clang 編譯，已測試的編譯器版本為：

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
    <td>Mac OS</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
  <tr>
    <td>XCode</td>
    <td>15.X</td>
    <td>x64、ARM64</td>
    <td>Sonoma 14.X</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
</table>

## 編譯過程

若使用命令列編譯，指令為
```
clang -fobjc-arc main.m
```
  
## 注意事項

本範例程式 Login 後返回之 token 其時效性為  
24 Hours。需在程式中實現重新 Login 的程式，  
這個功能並不包含在本範例程式之中。
