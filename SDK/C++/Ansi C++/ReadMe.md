 
## 說明

本程式採用標準 Ansi C++ 編譯，已測試的編譯器版本為：

<table>
  <tr>
    <td>編譯器</td>
    <td>版本</td>
    <td>編譯結果</td>
    <td>運作結果</td>
  </tr>
  <tr>
    <td>gcc</td>
    <td>13 (x86、x64)</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
  <tr>
    <td>gcc</td>
    <td>13 (ARM64)</td>
    <td>OK</td>
    <td>OK</td>
  </tr>
  <tr>
    <td>clang</td>
    <td>15 (ARM64)</td>
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
與 try catch 功能。所以需要 link curl、stdc++  
兩個 libraries。至於 pthread 等其餘函式庫，則  
看專案需求自行加入。  
