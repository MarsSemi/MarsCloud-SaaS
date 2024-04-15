 
## 說明

本程式採用標準 gcc 編譯，版本為 gcc 13。  

## 編譯過程

編譯指令為
```
gcc main.cpp -o test -lcurl -lstdc++
```

其中需要注意的是，C++ 版本的 SDK 會使用到 CURL  
與 try catch 功能。所以需要 link curl、stdc++  
兩個 libraries。至於 pthread 等其餘函式庫，則  
看專案需求自行加入。  
