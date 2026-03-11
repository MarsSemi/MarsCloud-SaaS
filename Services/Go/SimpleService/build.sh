#!/bin/bash

# 專案名稱
PROJECT_NAME="SimpleService"

echo "=== 開始編譯 ${PROJECT_NAME} ==="

# 整理 Go Modules
echo "正在執行 go mod tidy..."
go mod tidy

# 編譯執行檔
echo "正在編譯執行檔..."
go build -o ${PROJECT_NAME}

if [ $? -eq 0 ]; then
    echo "=== 編譯成功！執行檔名稱: ${PROJECT_NAME} ==="
else
    echo "=== 編譯失敗！ ==="
    exit 1
fi
