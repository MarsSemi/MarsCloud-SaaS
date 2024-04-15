 
## 說明

本程式採用標準 Ansi C++ 編譯，已測試的編譯器版本為：

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
    <td>gcc</td>
    <td>13</td>
    <td>x86、x64</td>
    <td>Ubuntu、Red Hat</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
  <tr>
    <td>gcc</td>
    <td>13</td>
    <td>ARM64</td>
    <td>Ubuntu、Mac OS</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
  <tr>
    <td>clang</td>
    <td>15</td>
    <td>ARM64</td>
    <td>Mac OS</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
</table>

## 編譯過程

編譯指令為
```
gcc main.cpp -o test -lcurl -lstdc++
```

其中需要注意的是，C++ 版本的 SDK 會使用到 CURL  
與 try catch 功能，所以需要 link curl、stdc++  
兩個 libraries。至於 pthread 等其餘函式庫，則  
看專案需求自行加入。  

## 注意事項

本範例程式 Login 後返回之 token 其時效性為  
24 Hours。需在程式中實現重新 Login 的程式，  
這個功能並不包含在本範例程式之中。
