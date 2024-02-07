#!/bin/bash
set -e

# 检查是否输入了标签
if [ $# -eq 0 ]; then
  echo "请提供版本标签，例如: ./build_and_push.sh v1.0.0"
  exit 1
fi

TAG=$1
echo "🚀 开始构建版本: $TAG"

# 为Git仓库添加标签
echo "📌 为Git仓库添加标签..."
git tag $TAG
git push origin $TAG
echo "✅ Git标签已创建并推送"

# 设置镜像名称
IMAGE_NAME="mangatal/bangumi-buddy"
LATEST_TAG="latest"

# 执行多平台构建并推送
echo "🔨 执行Docker多平台构建..."
docker buildx build --platform linux/amd64,linux/arm64 \
  -t $IMAGE_NAME:$LATEST_TAG \
  -t $IMAGE_NAME:$TAG \
  --push .

echo "✅ 构建完成并已推送至Docker Hub!"
echo "镜像标签:"
echo "  - $IMAGE_NAME:$LATEST_TAG"
echo "  - $IMAGE_NAME:$TAG" 