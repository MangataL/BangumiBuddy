#!/bin/bash
set -e

# 构建前端
echo "构建前端..."
cd webui
pnpm install
pnpm build
cd ..

# 创建后端web目录（如果不存在）
mkdir -p backend/web

# 复制前端构建产物到后端web目录
echo "复制前端资源到后端..."
cp -r webui/dist/* backend/web/

# 构建后端
echo "构建后端..."
cd backend
go build -o BangumiBuddy
cd ..

echo "构建完成！"
echo "可执行文件在: ./backend/BangumiBuddy" 